package main

import (
	"io"
	"log"
	"net/http"
	"sync"
)

type Client struct {
	name     string
	clientId string
	Message  chan string
}

type Message struct {
	name    string
	content string
}

var (
	broadcast chan Message

	clients []Client
	mu      sync.Mutex
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", nil))

	http.HandleFunc("/user", handleMessages)
}

func handleMessages(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	name := queryParams.Get("name")

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
	broadcast <- Message{name, string(body)}
	mu.Unlock()
}
