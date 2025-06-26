package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/client"
)

var (
	quit int
	err  error
)

func main() {
	quit = 0
	port := flag.Int("port", 8080, "HTTP Server Port")
	flag.Parse()

	url := fmt.Sprintf("http://localhost:%d", *port)

	client := client.NewClient()

	if err := client.Register(url); err != nil {
		log.Fatal(err)
	}

	//TODO quit channel zum stoppen nutzen!
	go func() {
		for quit == 0 {
			quit = client.GetMessages(url)
		}
	}()

	for quit == 0 {
		quit, err = client.PostMessage(url)
		if err != nil {
			log.Println("Fehler beim Absenden der Nachricht: ", err)
		}
	}
}
