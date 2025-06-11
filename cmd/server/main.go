package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	Name     string
	clientId string
}

type Message struct {
	name    string
	content string
}

var (
	broadcast chan Message      = make(chan Message)
	clients   map[string]Client = make(map[string]Client)
	mu        sync.Mutex
)

func main() {
	http.HandleFunc("/user/messages", handleMessages)
	http.HandleFunc("/user", handleRegistry)

	go func() {
		for {
			select {
			case msg := <-broadcast:
				sendBroadcast(msg)
			default:
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sendBroadcast(msg Message) {
	fmt.Printf("TODO")
}

func handleRegistry(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientId := queryParams.Get("name")

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
	clients[clientId] = Client{string(body), clientId}
	mu.Unlock()

}

func handleMessages(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientId := queryParams.Get("name")
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
	name := clients[clientId].Name
	broadcast <- Message{name, string(body)}
	mu.Unlock()
}
