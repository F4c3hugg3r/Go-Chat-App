package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/server"
)

func main() {
	var port = flag.Int("port", 8080, "HTTP Server Port")
	//var timeLimit = flag.Int("limit", 30, "Time limit for inactive clients in minutes")
	flag.Parse()
	portString := fmt.Sprintf(":%d", *port)

	service := server.NewChatService()
	plugin := server.RegisterPlugins(service)
	handler := server.NewServerHandler(service, plugin)

	http.HandleFunc("/users/{clientId}", handler.HandleMessages)
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
