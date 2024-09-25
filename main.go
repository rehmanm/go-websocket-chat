package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	UserName  string `json:"username"`
	Message   string `json:"message"`
	Broadcast bool   `json:"broadcast"`
	ToUser    string `json:"toUser"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Connected bool
	UserName  string
}

var clients = make(map[*websocket.Conn]Client)
var broadcast = make(chan Message)

func main() {

	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	fmt.Println("Server is running on port 8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error in starting server " + err.Error())
	}

}

func homePage(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Welcome to the Chatroom!")
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error in upgrading connection", err.Error())
		return
	}
	defer conn.Close()

	client := Client{Connected: true, UserName: "Anonymous"}

	clients[conn] = client

	for {
		var msg Message

		err := conn.ReadJSON(&msg)
		clients[conn] = Client{Connected: true, UserName: msg.UserName}
		if err != nil {
			fmt.Println("Error in reading message", err.Error())
			delete(clients, conn)
			return
		}

		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		for client := range clients {
			if clients[client].UserName == msg.UserName {
				continue
			}
			if (msg.Broadcast || msg.ToUser == clients[client].UserName) && clients[client].Connected {
				err := client.WriteJSON(msg)
				if err != nil {
					fmt.Println("Error in writing message", err.Error())
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}
