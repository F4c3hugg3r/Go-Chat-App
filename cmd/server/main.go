package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	server "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server"
)

// TODO Vorschläge
// Tests
// Abmelden
// Nachrichten an alle bei neuem user
// HTTPS
func main() {
	var port = flag.Int("port", 8080, "HTTP Server Port")
	flag.Parse()
	portString := fmt.Sprintf(":%d", *port)

	service := server.NewChatService()
	handler := server.NewServerHandler(service)

	http.HandleFunc("/users/{clientId}", handler.HandleRegistry)
	http.HandleFunc("/users/{clientId}/message", handler.AuthMiddleware(handler.HandleMessages))
	http.HandleFunc("/users/{clientId}/chat", handler.AuthMiddleware(handler.HandleGetRequest))

	go func() {
		for {
			time.Sleep(15 * time.Second)
			service.InactiveClientDeleter()
		}
	}()

	fmt.Println("Server running on port:", *port)
	log.Fatal(http.ListenAndServe(portString, nil))
}
