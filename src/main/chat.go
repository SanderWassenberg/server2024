package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	ws "github.com/gorilla/websocket"
	// https://pkg.go.dev/github.com/gorilla/websocket
	// how to websocket: https://www.golinuxcloud.com/golang-websocket/
)

var upgrader = ws.Upgrader{
	// ReadBufferSize:  1024,
	// WriteBufferSize: 1024,
}

type Chatrooms struct {
	Lock sync.Mutex // Lock when joining or leaving, that may remove or create a room
	Rooms map[string]Chatroom // room id is "lowId-highId"
}

type Chatroom struct {
	Lock sync.Mutex // lock when sending to a conn, and updating a conn
	Conn_1 *ws.Conn // lowest id
	Conn_2 *ws.Conn // highest id
}

/*
join(my_id, other_id, conn):
	create room id
	lock rooms

	get room ptr using the id
	if not present (
		create room
		add own conn
		add room
	) else (
		lock room
		add own conn
		unlock room
	)

	unlock rooms

leave(my_id, other_id):
	create room id
	lock rooms
	get room ptr using the id
	// it must be present

	lock room
	remove own conn
	if no other conn (
		remove room from rooms
	)
	unlock room

	unlock rooms

send():
	store msg in db
	lock room
	if other conn != nil (
		send over conn
	)
	unlock room
*/

func chat_handler(rw http.ResponseWriter, req *http.Request) {
	ok, info := check_auth(rw, req)
	if !ok { return }

	_ = info // stfu


	conn, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	go func () {
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
	} ()
	return
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

