package main

import (
	// "bytes"
	crypto_rand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	// "errors"
	// "fmt"
	// "io"
	"log"
	"net/http"
	// "os"
	// "os/signal"
	// "strconv"
	// "strings"
	"time"
	// "unicode/utf8"

	// "src/pwhash"
)



// base64 converts every 6 bits to a character.
// That means you get a 3:4 ratio, where every 3 bytes turn into 4 characters
// output length is always a multiple of 4.
// You can however strip padding characters
// bytes | padded len | stripped len
// 1		4		2
// 2		4		3
// 3		4		4
// 4		8		6

// ? ? ? ? ? ? ? ?,0 0 0 0 - - - -,- - - - - - - -
// [         ] [         ] [ padding ] [ padding ]

// ? ? ? ? ? ? ? ?,? ? ? ? ? ? ? ?,0 0 - - - - - -
// [         ] [         ] [         ] [ padding ]

// ? ? ? ? ? ? ? ?,? ? ? ? ? ? ? ?,? ? ? ? ? ? ? ?
// [         ] [         ] [         ] [         ]
// 48 bytes is a nice number, it results in a string of length 64
// 48 bytes = 2^384
// 24 bytes = 2^192
// 18 bytes = 2^144
// a uuid is 2^128, (16 bytes)
func generate_session_token() string {
	const n_bytes = 18
    buf := make([]byte, n_bytes)
	crypto_rand.Reader.Read(buf)
    return base64.StdEncoding.EncodeToString(buf) // It's bad to put session tokens in URL, so don't use URLEncoding.
}

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}


func login_handler(rw http.ResponseWriter, req *http.Request) {
	var ld LoginData
	if err := json.NewDecoder(req.Body).Decode(&ld); err != nil {
		log.Printf("login: %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if !check_login(ld.Username, ld.Password) {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	tok := generate_session_token()
	valid_until := time.Now().Add(1*time.Hour)

	if err := set_session(ld.Username, tok, valid_until); err != nil {
		log.Print(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}


	respond(rw, http.StatusOK, tok)

	// TODO: use set-cookie for this
}

func signup_handler(rw http.ResponseWriter, req *http.Request) {

}
