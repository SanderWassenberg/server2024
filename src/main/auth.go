package main

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	// "errors"
	// "fmt"
	// "io"
	// "log"
	// "net/http"
	// "os"
	// "os/signal"
	// "strings"
	// "unicode/utf8"
)



// base64 converts every 6 bits to a character.
// That means you get a 3:4 ratio, where every 3 bytes turn into 4 characters
// output length is always a multiple of 4
// 1 -> 4
// 2 -> 4
// 3 -> 4
// 4 -> 8

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
// a uuid is 2^128, which makes 18 bytes (24 chars) base64 string better. I think.
func rand_base64_string(n_bytes int) string {
    b := make([]byte, n_bytes)
	crypto_rand.Reader.Read(b)
    return base64.URLEncoding.EncodeToString(b) // could end up putting them in URLS, idk yet, might as well use this format since this is a web project.
}

