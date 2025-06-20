package main

import (
	"flag"
	"fmt"
	"log"

	service "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client"
)

var (
	quit error
)

func main() {
	quit = nil
	var port = flag.Int("port", 8080, "HTTP Server Port")
	flag.Parse()

	url := fmt.Sprintf("http://localhost:%d", *port)

	client := service.NewClient()

	if err := client.Register(url); err != nil {
		log.Fatal(err)
	}

	go func() {
		for quit == nil {
			client.GetMessages(url)
		}
	}()

	for quit == nil {
		quit = client.PostMessage(url)
	}
}
