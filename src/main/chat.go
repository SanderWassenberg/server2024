package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	ws "github.com/gorilla/websocket"
	// https://pkg.go.dev/github.com/gorilla/websocket
)

var upgrader = ws.Upgrader{
	// ReadBufferSize:  1024,
	// WriteBufferSize: 1024,
}

func chat_handler(rw http.ResponseWriter, req *http.Request) {

	// how to websocket: https://www.golinuxcloud.com/golang-websocket/

	// log.Print(req.Method)
	conn, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func () {
		conn.WriteMessage(ws.CloseMessage, nil) // if this errors, no problemo
		conn.Close()
	} ()

	for {
		// read a message
		messageType, messageContent, err := conn.ReadMessage()
		if err != nil {
			log.Printf("chat_handler readmsg: %v", err)
			return
		}

		// print out that message
		log.Printf("received: %v", string(messageContent))

		// reponse message
		messageResponse := fmt.Sprintf("Your message is: %s.", messageContent)

		if err := conn.WriteMessage(messageType, []byte(messageResponse)); err != nil {
			log.Printf("chat_handler writemsg: %v", err)
			return
		}
	}
}

func search_handler(rw http.ResponseWriter, req *http.Request) {
	ok, info := check_auth(rw, req)
	if !ok { return }

	searchterm, err := io.ReadAll(req.Body)
	results, err := search_by_interest(string(searchterm), info.Name, 100)

	if err != nil {
		log.Printf("search_handler: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	reslts_str, err := json.Marshal(results)
	if err != nil {
		log.Printf("search_handler: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	respond(rw, http.StatusOK, string(reslts_str))
}
func set_interest_handler(rw http.ResponseWriter, req *http.Request) {
	ok, info := check_auth(rw, req)
	if !ok { return }

	interest, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("set_interest: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = set_interest(info.Name, string(interest))
	if err != nil {
		log.Printf("set_interest: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func ban_handler(rw http.ResponseWriter, req *http.Request) {
	ok, info := check_auth(rw, req)
	if !ok { return }
	if info.Role != "admin" {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	who, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("ban_handler: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = set_banned(string(who), true)
	if err != nil {
		log.Printf("set_banned: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

