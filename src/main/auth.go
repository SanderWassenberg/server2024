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
	"unicode"
	// "unicode/utf8"
	"src/pwhash"

	"github.com/pquerna/otp/totp"
	"bytes"
	"io"
	"image/png"
)



// Types

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	OtpCode  string `json:"otp_code"`
}

type UserInfo struct {
	id       int    // Must remain lowercase. private = No json-mapping, we don't want to send this to users!
	Name     string `json:"name"`
	Role     string `json:"role"`
	Interest string `json:"interest"`
}



// Constants

const SessionTokenCookieName = "session_token"



// Utility functions

func generate_session_token() string {
	const n_bytes = 18
	buf := make([]byte, n_bytes)
	crypto_rand.Reader.Read(buf)
	return base64.StdEncoding.EncodeToString(buf) // It's bad to put session tokens in URL, so don't use URLEncoding.
}

func is_valid_username(s string) bool {
    for _, r := range s {
    	allowed := unicode.IsLetter(r) || r == '_'
        if !allowed { return false }
    }
    return true
}

func check_auth(rw http.ResponseWriter, req *http.Request) (authorized bool, info *UserInfo) {
	session_cookie, err := req.Cookie(SessionTokenCookieName)
	if err != nil {
		log.Printf("check_auth: %v", err)
		rw.WriteHeader(http.StatusUnauthorized)
		return false, nil
	}

	info, err = get_info_from_session(session_cookie.Value)
	if err != nil {
		log.Printf("check_auth: %v", err)
		if err == ErrSessionExpired || err == ErrSessionNotFound {
			respond(rw, http.StatusUnauthorized, err.Error())
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		return false, nil
	}

	return true, info
}



// Handlers

func login_handler(rw http.ResponseWriter, req *http.Request) {
	var ld LoginData
	if err := json.NewDecoder(req.Body).Decode(&ld); err != nil {
		log.Printf("login: %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	otp_info, err := get_otp_info(ld.Username)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("login_handler get_otp_info:", err)
		return
	}

	if otp_info.enabled == true {
		valid := totp.Validate(ld.OtpCode, otp_info.secret)
		if !valid {
			respond(rw, http.StatusUnauthorized, "Failed 2FA")
			log.Println("login_handler failed 2fa for user", ld.Username)
			return
		}
	}

	hash, err := get_password_hash(ld.Username)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("login_handler get_password_hash:", err)
		return
	}

	if !pwhash.VerifyPassword(ld.Password, hash) {
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

	if !is_valid_username(ld.Username) {
		respond(rw, http.StatusBadRequest, "Username may only contain letters and underscores.")
		return
	}

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

func otp_generate_handler(rw http.ResponseWriter, req *http.Request) {
	ok, info := check_auth(rw, req)
	if !ok {
		return
	}

	otp_info, err := get_otp_info(info.Name)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("otp_generate get_otp_info:", err)
		return
	}

	if otp_info.enabled == true {
		respond(rw, http.StatusBadRequest, "Unexpected request. 2FA already enabled on this account")
		log.Println("otp_generate unexpected: ", otp_info)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer: "sandershowcase.hbo-ict.org",
		AccountName: info.Name,
	})
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("otp_generate totp.Generate:", err)
		return
	}


	// Convert TOTP key into a QR code encoded as a PNG image.
	var buf bytes.Buffer
	buf.Write([]byte("data:text/plain;base64,"))
	base64encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	img, err := key.Image(200, 200)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("otp_generate key.image:", err)
		return
	}
	png.Encode(base64encoder, img)


	encoded_bytes, err := io.ReadAll(&buf)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("otp_generate io.readall:", err)
		return
	}

	err = set_otp_secret(info.Name, key.Secret())
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("otp_generate set_otp_secret:", err)
		return
	}

	respond(rw, http.StatusOK, string(encoded_bytes))
}

func otp_enable_handler(rw http.ResponseWriter, req *http.Request) {
	ok, info := check_auth(rw, req)
	if !ok {
		return
	}

	passcode, err := io.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("otp_enable io.readall:", err)
		return
	}

	otp_info, err := get_otp_info(info.Name)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		log.Println("otp_enable get_otp_info:", err)
		return
	}
	if otp_info.secret == "" || otp_info.enabled == true { // no secret, or already enabled
		respond(rw, http.StatusBadRequest, "Unexpected request. Server was not waiting for otp to be enabled on this account.")
		log.Println("otp_enable unexpected: ", otp_info)
		return
	}

	valid := totp.Validate(string(passcode), otp_info.secret)

	if !valid {
		respond(rw, http.StatusBadRequest, "incorrect otp passcode, ")
		log.Println("otp_enable failed to validate")
		return
	}

	set_otp_enabled(info.Name, true)

	rw.WriteHeader(http.StatusOK)
}