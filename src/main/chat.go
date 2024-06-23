package main

import (
	"fmt"
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
			log.Printf("readmsg err: %v", err)
			return
		}

		// print out that message
		log.Printf("received: %v", string(messageContent))

		// reponse message
		messageResponse := fmt.Sprintf("Your message is: %s.", messageContent)

		if err := conn.WriteMessage(messageType, []byte(messageResponse)); err != nil {
			log.Printf("writemsg err: %v", err)
			return
		}
	}
}