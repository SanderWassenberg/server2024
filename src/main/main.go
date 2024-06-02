package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	// "sync"
	// "sync/atomic"
	"log"
	"time"

	// "database/sql"
	// _ "github.com/mattn/go-sqlite3"
)

type Config struct {
	SendGrid_api_key string
	Port             string
}

var config Config

func main() {
	load_config_or_exit()
	if !strings.HasPrefix(config.Port, ":") { config.Port = ":" + config.Port } // ListenAndServe expects string to start with :

	// db, err := sql.Open("sqlite3", "showcase.db")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// defer func() { fmt.Println("Closing the db now!") }()
	// defer db.Close()

	file_server := http.FileServer(http.Dir("./static")) // this uses paths relative to the cwd of the exe

	// How to use Handle string pattern: https://pkg.go.dev/net/http@go1.22.3#ServeMux
	// NOTE: Most specific pattern takes precedence. Between "/" and "/api", the last is more specific, any url starting with "/api" will NOT go to the "/" handler.
	http.Handle("GET /", &PrintWrapper{file_server})
	http.HandleFunc("POST /api/contact", contact_func)
	http.HandleFunc("POST /api/wait_10", func (rw http.ResponseWriter, req *http.Request) {
		time.Sleep(60*time.Second)
		rw.WriteHeader(http.StatusOK)
	})

	// Graceful shutdown of ListenAndServe: https://stackoverflow.com/questions/39320025/how-to-stop-http-listenandserve
	// The waitgroup used in that example is unnecessary, Shutdown() will make ListenAndServe() return immediately, but will block *itself* until all open connections are closed.

    srv := &http.Server{Addr: config.Port}
	log.Printf("Starting server at port %v\n", config.Port)

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

	log.Println("Stopping HTTP server... This may take some time, we're waiting for all open connections to close.")
	if err := srv.Shutdown(ctx); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}

	log.Println("Exited gracefully.")
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



