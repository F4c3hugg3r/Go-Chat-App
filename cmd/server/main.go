package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Client is a communication participant who has a name, unique id and
// channel to receive messages
type Client struct {
	name     string
	clientId string
	clientCh chan Message
}

// Message contains the name of the sender and the message (content) itsself
type Message struct {
	name    string
	content string
}

var (
	clients map[string]Client = make(map[string]Client)
	mu      sync.Mutex
)

// TODO Flag definieren
func main() {
	var port = flag.Int("port", 8080, "HTTP Server Port")
	flag.Parse()
	portString := fmt.Sprintf(":%d", port)

	//Eigentlichg Query-Param aber dafür müsste ich externe bib nutzen
	http.HandleFunc("/user", handleRegistry)
	http.HandleFunc("/message", handleMessages)
	http.HandleFunc("/chat", handleGetRequest)

	log.Fatal(http.ListenAndServe(portString, nil))
}

// handleGetRequest displays a message when received and times out after 500s
// if nothing is being send
// should receive a Path Parameter with clientId in it eg "clientId?fgbIUHBVIUHDCdvw"
func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only GET Requests allowed", http.StatusBadRequest)
	}

	queryParams := r.URL.Query()
	clientId := queryParams.Get("clientId")
	if clientId == "" {
		http.Error(w, "missing query parameter", http.StatusBadRequest)
		return
	}

	select {
	case msg := <-clients[clientId].clientCh:
		message := msg.name + ": " + msg.content
		fmt.Fprint(w, message)
	case <-time.After(500 * time.Second):
		fmt.Println("No message received in time")
		return
	}
}

// handleRegistry lets a client register by it's name and id
// should receive a Path Parameter with clientId in it eg "clientId?fgbIUHBVIUHDCdvw"
// should receive the self given client-name in the request body
func handleRegistry(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientId := queryParams.Get("id")

	if r.Method != http.MethodPost {
		http.Error(w, "only POST request allowed", http.StatusBadRequest)
		return
	}

	if clientId == "" {
		http.Error(w, "missing query parameter", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	mu.Lock()
	if _, ok := clients[clientId]; ok {
		http.Error(w, "client already defined", http.StatusBadRequest)
		return
	}

	clientChan := make(chan Message)
	clients[clientId] = Client{string(body), clientId, clientChan}
	mu.Unlock()

	fmt.Printf("New client '%s' registered.", string(body))
}

// handleMessages takes an incoming message and distributes it to all clients
// should receive a Path Parameter with clientId in it eg "clientId?fgbIUHBVIUHDCdvw"
// should receive the message in the request body
func handleMessages(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientId := queryParams.Get("clientId")
	if clientId == "" {
		http.Error(w, "missing query parameter", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "only POST request allowed", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	name := clients[clientId].name
	sendBroadcast(Message{name, string(body)})
}

// sendBroadcast distributes an incomming message abroad all client Channels
func sendBroadcast(msg Message) {
	mu.Lock()
	defer mu.Unlock()
	for _, client := range clients {
		select {
		case client.clientCh <- msg:
		default:
			fmt.Printf("channel of client %s full, removing client", client.name)
			delete(clients, client.clientId)
		}
	}
}
