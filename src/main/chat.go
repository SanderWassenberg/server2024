package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	// "strings"

	// "time"

	ws "github.com/gorilla/websocket"
	// https://pkg.go.dev/github.com/gorilla/websocket
	// how to websocket: https://www.golinuxcloud.com/golang-websocket/
)

// Types

type Chatrooms struct {
	Lock sync.Mutex // Lock when joining or leaving, that may remove or create a room
	Rooms map[string]*Chatroom // room id is "lowId-highId"
}

type Chatroom struct {
	Lock sync.Mutex // lock when sending to a conn, and updating a conn
	Conn_1 *ws.Conn // lowest id
	Conn_2 *ws.Conn // highest id
}



// Vars

var upgrader = ws.Upgrader{
	// ReadBufferSize:  1024,
	// WriteBufferSize: 1024,

	// https://stackoverflow.com/questions/65034144/how-to-add-a-trusted-origin-to-gorilla-websockets-checkorigin
	CheckOrigin: func(r *http.Request) bool {
	    origin := r.Header.Get("Origin")
	    return origin == config.Trusted_Origin
	},
}

var chatrooms = Chatrooms{
	Rooms: make(map[string]*Chatroom),
}



// Utility functions

func my_sort(a, b int) (lo int, hi int) {
	if a < b {
		lo = a; hi = b
	} else {
		lo = b; hi = a
	}
	return
}

func join(my_id, other_id int, conn *ws.Conn) (room *Chatroom) {
	lo, hi := my_sort(my_id, other_id)
	room_id := fmt.Sprintf("%v-%v", lo, hi)
	room, ok := chatrooms.Rooms[room_id]
	if !ok {
		room = &Chatroom{}
		chatrooms.Rooms[room_id] = room
	}

	// lowest id is for conn_1
	if my_id < other_id {
		room.Conn_1 = conn
	} else {
		room.Conn_2 = conn
	}

	return room
}

func leave(my_id, other_id int) {
	lo, hi := my_sort(my_id, other_id)
	room_id := fmt.Sprintf("%v-%v", lo, hi)
	room, ok := chatrooms.Rooms[room_id]
	if !ok {
		log.Printf("room '%v' not present when it should be", room_id);
		panic("yikes!")
	}
	remove_room := false
	if my_id < other_id {
		room.Conn_1 = nil
		remove_room = room.Conn_2 == nil
	} else {
		room.Conn_2 = nil
		remove_room = room.Conn_1 == nil
	}

	if remove_room {
		chatrooms.Rooms[room_id] = nil
	}
}

func send(room *Chatroom, my_id, other_id int, text string) {
	err := save_message(my_id, other_id, text)
	if err != nil {
		log.Println("send() save_message:", err)
		// still continue
	}

	var send_conn *ws.Conn
	if my_id < other_id {
		send_conn = room.Conn_2
	} else {
		send_conn = room.Conn_1
	}

	if send_conn == nil { return }

	err = send_conn.WriteMessage(ws.TextMessage, []byte(text))
	if err != nil {
		log.Printf("send() writemsg: %v", err)
		return
	}
}



// Handlers

func chat_handler(rw http.ResponseWriter, req *http.Request) {
	ok, info := check_auth(rw, req)
	if !ok { return }

	// log.Println("this is the url:", req.URL)
	// log.Println("GET params were:", req.URL.Query())

	// https://stackoverflow.com/questions/15407719/in-gos-http-package-how-do-i-get-the-query-string-on-a-post-request
	other_username := req.URL.Query().Get("with")
	other_id, err := get_id(other_username)
	if err != nil {
		log.Printf("could not get id of other chatter '%v': %v", other_username, err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Println("websocket upgrade:", err)
		return
	}

	room := join(info.id, other_id, conn)

	go func () {
		defer func () {
			// conn.WriteMessage(ws.CloseMessage, nil) // if this errors, no problemo
			conn.Close()
		} ()

		for {
			// read a message
			messageType, messageContent, err := conn.ReadMessage()
			if err != nil {
				log.Printf("chat_handler readmsg: %v", err)
				return
			}

			if messageType == ws.CloseMessage {
				leave(info.id, other_id)
				return
			}

			send(room, info.id, other_id, string(messageContent))
		}
	} ()
	return
}

func chat_history_handler(rw http.ResponseWriter, req *http.Request) {
	ok, info := check_auth(rw, req)
	if !ok { return }

	other_username, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("chat_history_handler:", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	other_id, err := get_id(string(other_username))
	if err != nil {
		log.Println("chat_history_handler:", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	history, err := get_chat_history(info.id, other_id, 100)
	if err != nil {
		log.Println("chat_history_handler:", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	history_str, err := json.Marshal(history)
	if err != nil {
		log.Println("chat_history_handler:", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	respond(rw, http.StatusOK, string(history_str))
}

func search_handler(rw http.ResponseWriter, req *http.Request) {
	ok, info := check_auth(rw, req)
	if !ok { return }

	searchterm, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("search_handler:", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	results, err := search_by_interest(string(searchterm), info.id, 100)
	if err != nil {
		log.Println("search_handler:", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	results_str, err := json.Marshal(results)
	if err != nil {
		log.Println("search_handler:", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	respond(rw, http.StatusOK, string(results_str))
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

