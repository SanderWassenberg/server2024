package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

type PrintWrapper struct { http.Handler }

// atomic increment https://stackoverflow.com/questions/13908129/are-increment-operators-in-go-atomic-on-x86
var request_count atomic.Int64
func CustomPrint(req *http.Request) {
	request_count.Add(1) // atomic add
	fmt.Printf("--- Request %v ---\n", request_count.Load())
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
		fmt.Printf("Body:\n\"\"\"%v\"\"\"\n", string(body))
	} else {
		fmt.Printf("Body (Error from io.ReadAll): %v\n", err)
	}
	fmt.Println();
}


func (self *PrintWrapper) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	CustomPrint(req)
	self.Handler.ServeHTTP(rw, req)
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

type Contact_Data struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
	Email   string `json:"email"`
}

// returns utf8 length or "way too big" if the byte length is already more than 4x the limit
func verify_string_len(str string, lim int) (length any, ok bool) {
	if len(str) > lim * 4 { // utf8 uses at most 4 bytes per char, so more bytes than 4*limit is guaranteed to have more characters.
		return "way too big", false
	}
	l := utf8.RuneCountInString(str)
	return l, (l <= lim)
}

var email_regexp *regexp.Regexp

func contact_func(rw http.ResponseWriter, req *http.Request) {
	body, _ := io.ReadAll(req.Body)
	fmt.Printf("Body:\n\"\"\"%v\"\"\"\n", string(body))


	var contact_data Contact_Data
	if err := json.Unmarshal(body, &contact_data); err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		respond(rw, http.StatusBadRequest, err.Error());
		return;
	}
	fmt.Printf("contact data, parsed:\n%v\n", contact_data)

	if l, ok := verify_string_len(contact_data.Subject, 200); !ok {
		respond_fmt(rw, http.StatusBadRequest, "subject length should be <200, was %v", l)
		return
	}
	if l, ok := verify_string_len(contact_data.Message, 600); !ok {
		respond_fmt(rw, http.StatusBadRequest, "message length should be <600, was %v", l)
		return
	}
	if l, ok := verify_string_len(contact_data.Email, 100); !ok {
		respond_fmt(rw, http.StatusBadRequest, "email length should be <100, was %v", l)
		return
	}
	if ok := email_regexp.MatchString(contact_data.Email); !ok {
		respond(rw, http.StatusBadRequest, "invalid email")
		return
	}

	respond(rw, http.StatusOK, "ayo we mailed that shit real good (not really)");
	time.Sleep(2 * time.Second)
}

func main() {
	// See:
	// https://stackoverflow.com/questions/201323/how-can-i-validate-an-email-address-using-a-regular-expression
	// https://regexper.com/#(%3F%3A%5Ba-z0-9!%23%24%25%26%27*%2B%2F%3D%3F%5E_%60%7B%7C%7D%7E-%5D%2B(%3F%3A%5C.%5Ba-z0-9!%23%24%25%26%27*%2B%2F%3D%3F%5E_%60%7B%7C%7D%7E-%5D%2B)*%7C%22(%3F%3A%5B%5Cx01-%5Cx08%5Cx0b%5Cx0c%5Cx0e-%5Cx1f%5Cx21%5Cx23-%5Cx5b%5Cx5d-%5Cx7f%5D%7C%5C%5C%5B%5Cx01-%5Cx09%5Cx0b%5Cx0c%5Cx0e-%5Cx7f%5D)*%22)%40(%3F%3A(%3F%3A%5Ba-z0-9%5D(%3F%3A%5Ba-z0-9-%5D*%5Ba-z0-9%5D)%3F%5C.)%2B%5Ba-z0-9%5D(%3F%3A%5Ba-z0-9-%5D*%5Ba-z0-9%5D)%3F%7C%5C%5B(%3F%3A(%3F%3A(2(5%5B0-5%5D%7C%5B0-4%5D%5B0-9%5D)%7C1%5B0-9%5D%5B0-9%5D%7C%5B1-9%5D%3F%5B0-9%5D))%5C.)%7B3%7D(%3F%3A(2(5%5B0-5%5D%7C%5B0-4%5D%5B0-9%5D)%7C1%5B0-9%5D%5B0-9%5D%7C%5B1-9%5D%3F%5B0-9%5D)%7C%5Ba-z0-9-%5D*%5Ba-z0-9%5D%3A(%3F%3A%5B%5Cx01-%5Cx08%5Cx0b%5Cx0c%5Cx0e-%5Cx1f%5Cx21-%5Cx5a%5Cx53-%5Cx7f%5D%7C%5C%5C%5B%5Cx01-%5Cx09%5Cx0b%5Cx0c%5Cx0e-%5Cx7f%5D)%2B)%5C%5D)
	// to ensure the whole string is matched and not a portion: https://stackoverflow.com/questions/447250/matching-exact-string-with-javascript
	email_regexp = regexp.MustCompile(`^(?:[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$`);

	file_server := http.FileServer(http.Dir("./static")) // this uses paths relative to the cwd of the exe

	// How to use Handle string pattern: https://pkg.go.dev/net/http@go1.22.3#ServeMux
	// NOTE: Most specific pattern takes precedence. Between "/" and "/api", the last is more specific, any url starting with "/api" will NOT go to the "/" handler.
	http.Handle("GET /", &PrintWrapper{file_server})
	http.HandleFunc("POST /api/contact", contact_func)


	port := ":8080" // the prefixed : is required for ListenAndServe
	fmt.Printf("Starting server at port %v\n", port)

	// if err := http.ListenAndServe(port, nil); err != nil {
	// 	log.Fatal(err)
	// }
	log.Fatal(http.ListenAndServe(port, nil))
}