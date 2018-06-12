package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
)

var clients = make(map[*websocket.Conn]bool) //connected clinets
var broadcast = make(chan Message) //broadcast channel

var upgrader = websocket.Upgrader{}

// Define our message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	fs:=http.FileServer(http.Dir("../public"))
	http.Handle("/",fs)

	//configure websocket route
	http.HandleFunc("/ws", handleConnections)

	//// Start listening for incoming chat messages
	go handleMessage()
	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func handleConnections(w http.ResponseWriter,r *http.Request){
	//// Upgrade initial GET request to a websocket
	ws,err := upgrader.Upgrade(w,r,nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	// register out new client
	clients[ws] = true

	for  {
		//// Read in a new message as JSON and map it to a Message object
		var msg Message  //read in a new
		err := ws.ReadJSON(&msg)
		if err != nil{
			log.Printf("error:%v",err)
			delete(clients,ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}
func handleMessage(){
	for {
		msg := <- broadcast
		for client := range clients{
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error:%v",err)
				client.Close()
				delete(clients,client)
			}
		}
	}
}