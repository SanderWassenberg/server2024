package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"unicode/utf8"
)

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

// returns utf8 length or "way too big" if the byte length is already more than 4x the limit
func verify_utf8_string_len(str string, lim int) (length any, ok bool) {
	if len(str) > lim * 4 { // utf8 uses at most 4 bytes per char, so more bytes than 4*limit is guaranteed to have more characters.
		return "way too big", false
	}
	l := utf8.RuneCountInString(str)
	return l, (l <= lim)
}

type MyError struct {
	msg string
	args []any
}
type MyErrors []*MyError

func my_error(msg string, args ...any) *MyError {
	return &MyError{msg, args}
}
func (self *MyError) Error() string {
	if len(self.args) == 0 { return self.msg }
	return fmt.Sprintf(self.msg, self.args...)
}
func (self *MyErrors) Error() string {
	b := strings.Builder{}
	for _, v := range *self {
		fmt.Fprintf(&b, v.msg, v.args...)
		fmt.Fprint(&b, "\n")
	}
	return b.String()
}

func wait_until_interrupt() {
	// Intercept or wait for Ctrl-C (SIGINT): https://stackoverflow.com/questions/11268943/is-it-possible-to-capture-a-ctrlc-signal-sigint-and-run-a-cleanup-function-i
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop // blocks until the channel receives a value
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