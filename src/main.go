package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
)

type PrintWrapper struct { http.Handler }

// atomic increment https://stackoverflow.com/questions/13908129/are-increment-operators-in-go-atomic-on-x86
var request_count atomic.Int64
func CustomPrint(req *http.Request) {
	request_count.Add(1) // atomic add
	fmt.Printf("\n--- Request %v ---\n", request_count.Load())
	fmt.Printf("Method: \"%v\"\n", req.Method)
	fmt.Printf("URL: \"%v\"\n", req.URL)
	fmt.Print("Headers:\n")
	for headername, headervalues := range req.Header {
		if len(headervalues) == 1 {
			fmt.Printf("- %v: \"%v\"\n", headername, headervalues[0])
		} else {
			fmt.Printf("- %v: Included %v times: [", headername, len(headervalues))
			for _, value := range headervalues { fmt.Printf("\"%v\"", value) }
			fmt.Print("]\n")
		}
	}
	if body, err := io.ReadAll(req.Body); err == nil {
		fmt.Printf("Body:\n\"\"\"%v\"\"\"\n", body)
	} else {
		fmt.Printf("Body (Error from io.ReadAll): %v\n", err)
	}
}


func (self *PrintWrapper) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	CustomPrint(req)
	self.Handler.ServeHTTP(rw, req)
}

func main() {
	file_server := http.FileServer(http.Dir("./static")) // this uses paths relative to the cwd of the exe

	http.Handle("/", &PrintWrapper{file_server})
	// http.Handle("/api/", &myserv)
	// most specific pattern takes precedence so "/api" will get all requests to "/api", and "/" will not. See https://pkg.go.dev/net/http@go1.22.3#ServeMux

	port := ":8080" // the prefixed : is required for ListenAndServe

	fmt.Printf("Starting server at port %v\n", port)

	// if err := http.ListenAndServe(port, nil); err != nil {
	// 	log.Fatal(err)
	// }
	log.Fatal(http.ListenAndServe(port, nil))
}