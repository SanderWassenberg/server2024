package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)



// Types

type LogWrapper struct {
	http.Handler
	Static bool
}



// Utility functions

func Wrap(f func(http.ResponseWriter, *http.Request)) http.Handler {
	return &LogWrapper{ Handler: http.HandlerFunc(f), Static: false }
}



// Methods

func (self *LogWrapper) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	self.CustomPrint(req)
	self.Handler.ServeHTTP(rw, req)
}

func (self *LogWrapper) CustomPrint(req *http.Request) {
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
