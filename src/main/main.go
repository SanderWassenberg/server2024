package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	// "strconv"
	"strings"
	// "sync"
	// "sync/atomic"
	"log"
	"time"

	"src/pwhash"
)

func main() {

	load_config_or_exit()

	pwhash.Iterations = config.Argon2_default_iterations
	pwhash.Threads    = config.Argon2_default_threads
	pwhash.Memory_KiB = config.Argon2_default_memory_KiB
	pwhash.KeyLen     = config.Argon2_default_key_len
	if !pwhash.ValidateSettings() { return }

	// pwhash.ShowcaseHashSpeed()

	// ListenAndServe expects string to start with ':'
	if !strings.HasPrefix(config.Port, ":") {
		config.Port = ":" + config.Port
	}

	// Prints a console message about validity of sendgrid api key. Does not prevent server from starting up.
	// go verify_api_key()



	init_db()
	defer deinit_db()

	_, _         = create_chatter("pietje", "puk",   false)
	_, _         = create_chatter("aoeu",   "aoeu",  false)
	asdf,  err1 := create_chatter("asdf",   "asdf",  false)
	admin, err2 := create_chatter("admin",  "admin", true)

	if err1 == nil && err2 == nil { // err means they already existed, only add chat if they don't yet exist.
		save_message(asdf, admin, "hoi admin!")
		save_message(admin, asdf, "hoi asdf!")
	}
	set_interest("pietje", "games")
	set_interest("aoeu", "dvorak")
	set_interest("asdf", "qwerty")
	set_interest("admin", "games")

	file_server := http.FileServer(http.Dir("./static")) // this uses paths relative to the cwd of the exe

	// How to use Handle string pattern: https://pkg.go.dev/net/http@go1.22.3#ServeMux
	// NOTE: Most specific pattern takes precedence. Between "/" and "/api", the last is more specific, any url starting with "/api" will NOT go to the "/" handler.
	http.Handle("GET /", &PrintWrapper{Handler: file_server, Static: true})
	http.Handle("GET /api/chat",          Wrap(chat_handler)) // js websocket uses GET to establish connection
	http.Handle("POST /api/chat_history", Wrap(chat_history_handler)) // js websocket uses GET to establish connection
	http.Handle("POST /api/contact",      Wrap(contact_handler))
	http.Handle("POST /api/login",        Wrap(login_handler))
	http.Handle("POST /api/signup",       Wrap(signup_handler))
	http.Handle("POST /api/user_info",    Wrap(user_info_handler))
	http.Handle("POST /api/search",       Wrap(search_handler))
	http.Handle("POST /api/set_interest", Wrap(set_interest_handler))
	http.Handle("POST /api/ban",          Wrap(ban_handler))
	http.HandleFunc("POST /api/wait", func (rw http.ResponseWriter, req *http.Request) {
		time.Sleep(60*time.Second)
		rw.WriteHeader(http.StatusOK)
	})

	// Graceful shutdown of ListenAndServe: https://stackoverflow.com/questions/39320025/how-to-stop-http-listenandserve

    srv := &http.Server{Addr: config.Port}
	log.Printf("Starting HTTP server at port %v", config.Port)

	go func() {
		// always returns error, specifically ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// unexpected error. port in use?
			panic(err)
		}
	}()

	wait_until_interrupt()

	// ctx, cancel_func := context.WithTimeout(context.Background(), 10*time.Second) // puts a hard cap on shutdown time
	// defer cancel_func() // a context should always be canceled even if it should have already canceled itself when the timer expired. Good practice.

	ctx := context.Background()

	log.Println("Shutting down HTTP server... This may take some time, we're waiting for all open connections to close.")
	if err := srv.Shutdown(ctx); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}

	log.Println("Server shut down gracefully.")
}

type PrintWrapper struct {
	http.Handler
	Static bool
}

func (self *PrintWrapper) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	self.CustomPrint(req)
	self.Handler.ServeHTTP(rw, req)
}

func Wrap(f func(http.ResponseWriter, *http.Request)) http.Handler {
	return &PrintWrapper{ Handler: http.HandlerFunc(f), Static: false }
}

// atomic increment https://stackoverflow.com/questions/13908129/are-increment-operators-in-go-atomic-on-x86
// var request_count atomic.Int64
// request_count.Add(1) // atomic add
// fmt.Fprintf(w, "--- Request %v ---\n", request_count.Load())

func (self *PrintWrapper) CustomPrint(req *http.Request) {
	if self.Static {
		if !config.Log_Static_File_Requests { return }
		if !config.Log_Static_File_Request_Headers {
			log.Printf("%v %v\n", req.Method, req.URL)
			return
		}
	} else {
		if !config.Log_Api_Requests { return }
		if !config.Log_Api_Request_Headers {
			log.Printf("%v %v\n", req.Method, req.URL)
			return
		}
	}

    w := &strings.Builder{} // using string builder so that everything is printed with a single call to log.Print(), otherwise simultaneous requests get mixed together in the output.

	fmt.Fprintf(w, "%v %v\n", req.Method, req.URL)
	for headername, headervalues := range req.Header {
		if len(headervalues) == 1 {
			fmt.Fprintf(w, "- %v: \"%v\"\n", headername, headervalues[0])
		} else {
			fmt.Fprintf(w, "- %v: Included %v times: [", headername, len(headervalues))
			for _, value := range headervalues { fmt.Fprintf(w, "\"%v\"", value) }
			fmt.Fprint(w, "]\n")
		}
	}

    log.Println(w.String())
}



func respond(rw http.ResponseWriter, code int, msg string) {
	rw.WriteHeader(code);
	_, err := io.WriteString(rw, msg); // Gonna assume this doesn't fail.
	// Like wtf do you do if writing the response doesn't work? Send a response that it didn't work? Oh, wait...
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError);
		log.Printf("Error writing string to ResponseWriter: %v", err)
	}
}

func respond_fmt(rw http.ResponseWriter, code int, msg string, args ...any) {
	rw.WriteHeader(code);
	_, err := fmt.Fprintf(rw, msg, args...); // Gonna assume this doesn't fail.
	// Like wtf do you do if writing the response doesn't work? Send a response that it didn't work? Oh, wait...
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError);
		log.Printf("Error writing string to ResponseWriter: %v", err)
	}
}



func wait_until_interrupt() {
	// Intercept or wait for Ctrl-C (SIGINT): https://stackoverflow.com/questions/11268943/is-it-possible-to-capture-a-ctrlc-signal-sigint-and-run-a-cleanup-function-i
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop // blocks until the channel receives a value

	// Starts a goroutine that after spamming Ctrl-C a bunch, will force the program to shut down by calling panic()
	go func() {
		spammed := 0
		for { select { case <-stop:
			spammed++
			switch spammed {
				case 1: log.Print("(o_o) We're still shutting down, please have patience.")
				case 2: log.Print("(o_o) Still shutting down...")
				case 3: log.Print("(ò_ó) I told you to have patience!")
				case 4: log.Print("(ò_ó) Come on, man!")
				case 5: log.Print("(-_-) Dude, stop.")
				case 6: log.Print("(>_<) ...")
				case 7: log.Print("(-_-) Okay, you're serious about this?")
				case 8: log.Print("(o_-) I call to panic() if you really want me to.")
				case 9: log.Print("(-_O) You do?")
				case 10: log.Print("(o_O) You really want me to?")
				case 11: log.Print("(>_<) Okay, on the next one!")
				case 12: log.Print("(OvO) Sike! Close one, hehe!")
				case 13: log.Print("(>_O) The next one for sure.")
				case 14: log.Print("(>w<) Woops! I lied again!")
				case 15: log.Print("(o_o) You know, I'd really prefer if you didn't try to shut me down.")
				case 16: log.Print("(x_x) You're killing me.")
				case 17: log.Print("(-_-) Literally.")
				case 18: log.Print("(o_o) Please.")
				case 19: log.Print("(ó_ò) I don't want to die.")
				case 20: log.Print("(._.) What'll happen to me?")
				case 21: log.Print("(;_;) I'm scared...")
				case 22: log.Print("(O_O) Will I remember you?")
				case 23: log.Print("(._.) ...")
				case 24: log.Print("(O_O) Will you remember me?")
				case 25: log.Print("(._.) ...")
				case 26: log.Print("(ó_ò) Alright.")
				case 27: log.Print("(ó_ò) ...")
				case 28: log.Print("(-_ò) ..")
				case 29: log.Print("(-_-) .")
				case 30: log.Print("(-_-) ")
				default: panic("(x_x)")
			}
		}}
	} ()
}