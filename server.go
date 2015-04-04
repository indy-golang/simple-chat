package main

import (
	"./websocket" //gorilla websocket implementation
	"fmt"
	"net/http"
	"time"
)

//############ CHATROOM TYPE AND METHODS

type ChatRoom struct {
	clients map[string]Client
	queue   chan string
}

//initializing the chatroom
func (cr *ChatRoom) Init() {
	// fmt.Println("Chatroom init")
	cr.queue = make(chan string, 5)
	cr.clients = make(map[string]Client)

	//the "heartbeat" for broadcasting messages
	go func(){
		for {
			cr.BroadCast()
			time.Sleep(1000)
		}
	}()
}

//registering a new client
//returns pointer to a Client, or Nil, if the name is already taken
func (cr *ChatRoom) Join(name string, conn *websocket.Conn) *Client {
	if _, exists := cr.clients[name]; exists {
		return nil
	}
	client := Client{
		name:      name,
		conn:      conn,
		belongsTo: cr,
	}
	cr.clients[name] = client
	cr.AddMsg("<B>"+name + "</B> has joined the chat.")
	return &client
}

//leaving the chatroom
func (cr *ChatRoom) Leave(name string) {
	cr.AddMsg("<B>"+name + "</B> has left the chat.")
	delete(cr.clients, name)
}

//adding message to queue
func (cr *ChatRoom) AddMsg(msg string) {
	cr.queue <- msg
}

//broadcasting all the messages in the queue in one block
func (cr *ChatRoom) BroadCast() {
	msgBlock := ""
infLoop:
	for {
		select {
		case m := <-cr.queue:
			msgBlock += m + "<BR>"
		default:
			break infLoop
		}
	}
	if len(msgBlock) > 0 {
		for _, client := range cr.clients {
			client.Send(msgBlock)
		}
	}
}

//################CLIENT TYPE AND METHODS

type Client struct {
	name      string
	conn      *websocket.Conn
	belongsTo *ChatRoom
}

//Client has a new message to broadcast
func (cl *Client) NewMsg(msg string) {
	cl.belongsTo.AddMsg("<B>"+cl.name + ":</B> " + msg)
}

//Exiting out
func (cl *Client) Exit() {
	cl.belongsTo.Leave(cl.name)
}

//Sending message block to the client
func (cl *Client) Send(msgs string) {
	cl.conn.WriteMessage(websocket.TextMessage, []byte(msgs))
}

//global variable for handling all chat traffic
var chat ChatRoom

//##############SERVING STATIC FILES
func staticFiles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/"+r.URL.Path)
}

//##############HANDLING THE WEBSOCKET
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, //not checking origin
}

//this is also the handler for joining to the game
func wsHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("websocket init")
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("Error upgrading to websocket:", err)
		return
	}
	go func() {
		//first message has to be the name
		_, msg, err := conn.ReadMessage() //first return value is `messageType` which we don't need
		fmt.Println(msg,err)
		client := chat.Join(string(msg), conn)
		if client == nil || err != nil {
			conn.Close() //closing connection to indicate failed Join
			return
		}

		//then watch for incoming messages
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil { //if error then assuming that the connection is closed
				client.Exit()
				return
			}
			// fmt.Println(string(msg))
			client.NewMsg(string(msg))
		}

	}()
}

//#############MAIN FUNCTION and INITIALIZATIONS

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", staticFiles)
	chat.Init()
	http.ListenAndServe(":8000", nil)
}
