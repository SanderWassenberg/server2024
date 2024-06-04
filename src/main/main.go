package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	// "sync"
	// "sync/atomic"
	"log"
	"time"

	"database/sql"
	_ "github.com/mattn/go-sqlite3" // calls package init func which adds sqlite3 as a driver for database/sql to use
)



func main() {
	// todo, follow auth tutorial
	// https://www.sohamkamani.com/golang/session-cookie-authentication/
	// using base64 rand string as session token

	load_config_or_exit()

	// ListenAndServe expects string to start with ':'
	if !strings.HasPrefix(config.Port, ":") {
		config.Port = ":" + config.Port
	}

	// Prints a console message about validity of sendgrid api key. Does not prevent server from starting up.
	go verify_api_key()

	db, err := sql.Open("sqlite3", "showcase.db")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func () {
		log.Print("Closing db...")
		if err := db.Close(); err != nil {
			log.Printf("Error closing db: %v", err)
			return
		}
		log.Print("Successfully closed db.")
	} ()


	// Create table
// 	statement, err := db.Prepare(`
// CREATE TABLE IF NOT EXISTS books (
// 	id INTEGER PRIMARY KEY,
// 	isbn INTEGER,
// 	author TEXT,
// 	name TEXT NULL
// ) STRICT
// `)

// 	if err != nil {
// 		log.Println("Error in creating table")
// 	} else {
// 		log.Println("Successfully created table books with strict stuff or whatever!")
// 	}
// 	statement.Exec()





	file_server := http.FileServer(http.Dir("./static")) // this uses paths relative to the cwd of the exe

	// How to use Handle string pattern: https://pkg.go.dev/net/http@go1.22.3#ServeMux
	// NOTE: Most specific pattern takes precedence. Between "/" and "/api", the last is more specific, any url starting with "/api" will NOT go to the "/" handler.
	http.Handle("GET /", &PrintWrapper{file_server})
	http.HandleFunc("POST /api/contact", contact_func)
	http.HandleFunc("POST /api/wait", func (rw http.ResponseWriter, req *http.Request) {
		time.Sleep(60*time.Second)
		rw.WriteHeader(http.StatusOK)
	})

	// Graceful shutdown of ListenAndServe: https://stackoverflow.com/questions/39320025/how-to-stop-http-listenandserve

    srv := &http.Server{Addr: config.Port}
	log.Printf("Starting HTTP server at port %v", config.Port)

	go func() {
		// always returns error. ErrServerClosed on graceful close
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



type PrintWrapper struct { http.Handler }

func (self *PrintWrapper) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	CustomPrint(req)
	self.Handler.ServeHTTP(rw, req)
}

// atomic increment https://stackoverflow.com/questions/13908129/are-increment-operators-in-go-atomic-on-x86
// var request_count atomic.Int64
// request_count.Add(1) // atomic add
// fmt.Fprintf(w, "--- Request %v ---\n", request_count.Load())

func CustomPrint(req *http.Request) {
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

	if body, err := io.ReadAll(req.Body); err == nil {
		if len(body) == 0 {
			fmt.Fprint(w, "Body: <empty>\n")
		} else {
			fmt.Fprintf(w, "Body:\n`%v`\n", string(body))
		}
	} else {
		fmt.Fprintf(w, "Body: (Error from io.ReadAll) %v\n", err)
	}

    log.Println(w.String())
}



func respond(rw http.ResponseWriter, code int, msg string) {
	rw.WriteHeader(code);
	_, _ = io.WriteString(rw, msg); // Gonna assume this doesn't fail.
	// Like wtf do you do if writing the response doesn't work? Send a response that it didn't work? Oh, wait...
}

func respond_fmt(rw http.ResponseWriter, code int, msg string, args ...any) {
	rw.WriteHeader(code);
	_, _ = fmt.Fprintf(rw, msg, args...); // Gonna assume this doesn't fail.
	// Like wtf do you do if writing the response doesn't work? Send a response that it didn't work? Oh, wait...
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
				case 1: log.Print("(o_o) We're still shutting down, please have patience")
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
				case 14: log.Print("(>w<) Woops! I lied again")
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