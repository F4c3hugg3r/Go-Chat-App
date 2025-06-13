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
	quit     chan int
	active   bool
}

// Message contains the name of the sender and the message (content) itsself
type Message struct {
	name    string
	content string
}

var (
	clients map[string]*Client = make(map[string]*Client)
	mu      sync.RWMutex
)

func main() {
	var port = flag.Int("port", 8080, "HTTP Server Port")
	flag.Parse()
	portString := fmt.Sprintf(":%d", *port)

	//Eigentlichg Query-Param aber dafür müsste ich externe bib nutzen
	http.HandleFunc("/user", handleRegistry)
	http.HandleFunc("/message", handleMessages)
	http.HandleFunc("/chat", handleGetRequest)

	go func() {
		for {
			time.Sleep(15 * time.Second)
			inactiveClientDeleter()
		}
	}()

	fmt.Println("Server running on port:", *port)
	log.Fatal(http.ListenAndServe(portString, nil))
}

func inactiveClientDeleter() {
	mu.Lock()
	defer mu.Unlock()

	for clientId, client := range clients {
		if !client.active {
			fmt.Println("due to inactivity: deleting ", client.name)
			delete(clients, clientId)
		}
	}
}

// handleRegistry lets a client register by it's name and id
// should receive a Path Parameter with clientId in it eg "clientId?fgbIUHBVIUHDCdvw"
// should receive the self given client-name in the request body
func handleRegistry(w http.ResponseWriter, r *http.Request) {
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
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	err2 := registerClient(clientId, string(body))
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusBadRequest)
		return
	}
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
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}

	mu.RLock()
	name := clients[clientId].name
	mu.RUnlock()

	sendBroadcast(Message{name, string(body)})
}

// sendBroadcast distributes an incomming message abroad all client channels
func sendBroadcast(msg Message) {
	mu.Lock()
	defer mu.Unlock()
	if len(clients) <= 0 {
		fmt.Printf("\nThere are no clients registered\n")
		return
	}

	for _, client := range clients {
		select {
		case client.clientCh <- msg:
			fmt.Println("success")
		default:
			client.active = false
		}
	}
}

//TODO nur ein User darf auf einen Channel hören

// handleGetRequest displays a message when received and times out after 250s
// if nothing is being send
// should receive a Path Parameter with clientId in it eg "clientId?fgbIUHBVIUHDCdvw"
func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only GET Requests allowed", http.StatusBadRequest)
		return
	}

	queryParams := r.URL.Query()
	clientId := queryParams.Get("clientId")
	if clientId == "" {
		http.Error(w, "missing query parameter", http.StatusBadRequest)
		return
	}

	mu.RLock()
	client, ok := clients[clientId]
	isActive := ok && client.active
	mu.RUnlock()

	if !isActive {
		http.Error(w, "Client not found or not active", http.StatusNotFound)
		return
	}

	select {
	case msg := <-client.clientCh:
		message := msg.name + ": " + msg.content
		fmt.Fprint(w, message)
		return
	case <-client.quit:
		mu.Lock()
		client.active = false
		close(client.clientCh)
		close(client.quit)
		mu.Unlock()
		return
	case <-time.After(250 * time.Second):
		return
	}
}

func registerClient(clientId, body string) error {
	mu.Lock()
	if _, ok := clients[clientId]; ok {
		return fmt.Errorf("client already defined")
	}
	clientCh := make(chan Message)
	quit := make(chan int)
	clients[clientId] = &Client{string(body), clientId, clientCh, quit, true}
	mu.Unlock()
	fmt.Printf("\nNew client '%s' registered.\n", body)
	return nil
}
