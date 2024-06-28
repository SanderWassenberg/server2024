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

const SessionTokenCookieName = "session_token"

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

// type LoginResponse struct {
// 	SessionToken string `json:"sessionToken"`
// 	Role string `json:"role"`
// }

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

	session_token := generate_session_token()
	valid_until := time.Now().Add(1 * time.Hour)

	if err := set_session(ld.Username, session_token, valid_until); err != nil {
		log.Printf("login: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// respond(rw, http.StatusOK, json.Marshall(response))

	http.SetCookie(rw, &http.Cookie{
		Name:    SessionTokenCookieName,
		Value:   session_token,
		Expires: valid_until,
	})

	rw.WriteHeader(http.StatusOK)
}

func signup_handler(rw http.ResponseWriter, req *http.Request) {
	var ld LoginData
	if err := json.NewDecoder(req.Body).Decode(&ld); err != nil {
		log.Println("signup_handler:", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// should probably not allow all strings as usernames,
	// currently you can have usernames with spaces and special characters and shit and that's probably not good.

	_, err := create_chatter(ld.Username, ld.Password, false)
	if err != nil {
		log.Println("signup handler:", err)
		if err == ErrUserAlreadyExists {
			respond(rw, http.StatusBadRequest, err.Error())
		} else {
			rw.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func user_info_handler(rw http.ResponseWriter, req *http.Request) {
	ok, info := check_auth(rw, req)
	if !ok {
		return
	}

	info_json, err := json.Marshal(info)
	if err != nil {
		log.Println("user_info:", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	respond(rw, http.StatusOK, string(info_json))
}

type UserInfo struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Interest string `json:"interest"`
}

func check_auth(rw http.ResponseWriter, req *http.Request) (authorized bool, info *UserInfo) {
	session_cookie, err := req.Cookie(SessionTokenCookieName)
	if err != nil {
		log.Printf("check_auth: %v", err)
		rw.WriteHeader(http.StatusUnauthorized)
		return false, nil
	}

	var name string
	name, err = get_name_from_session(session_cookie.Value)
	if err != nil {
		log.Printf("check_auth: %v", err)
		if err == ErrSessionExpired || err == ErrSessionNotFound {
			respond(rw, http.StatusUnauthorized, err.Error())
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		return false, nil
	}

	isadmin, err := is_admin(name)
	if err != nil {
		log.Printf("check_auth: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return false, nil
	}

	var role string
	if isadmin {
		role = "admin"
	} else {
		role = "chatter"
	}

	interest, err := get_interest(name)
	if err != nil {
		log.Printf("check_auth: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return false, nil
	}

	return true, &UserInfo{
		Role:     role,
		Name:     name,
		Interest: interest,
	}
}
