package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/F4c3hugg3r/Go-Chat-Server/pkg/client"
)

type Config struct {
	port int
}

func main() {
	cfg := NewConfig()
	wg := &sync.WaitGroup{}
	url := fmt.Sprintf("http://localhost:%d", cfg.port)
	client := client.NewClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interChan := make(chan os.Signal, 2)
	signal.Notify(interChan, os.Interrupt, syscall.SIGTERM)

	wg.Add(1)

	go interruptListener(interChan, cancel, wg, client, url, ctx)

	err := client.Register(url)
	if err != nil {
		log.Fatal(err)
	}

	startChat(wg, ctx, client, url, cancel)

	wg.Wait()
}

// interruptListener sends a cancel() signal and closes all connections and requests if a interruption like
// os.Interrupt or syscall.SIGTERM is being triggered
func interruptListener(interChan chan os.Signal, cancel context.CancelFunc, wg *sync.WaitGroup, client *client.Client, url string, ctx context.Context) {
	defer wg.Done()

	select {
	case <-interChan:
		client.PostMessage(url, cancel, "/quit")
		cancel()
	case <-ctx.Done():
	}

	client.HttpClient.CloseIdleConnections()

	log.Println("Client logged out")
}

// startChat starts two go-routines for the sending and receiving of messages/responses
func startChat(wg *sync.WaitGroup, ctx context.Context, client *client.Client, url string, cancel context.CancelFunc) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				client.GetMessages(url, cancel)
			}
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := client.PostMessage(url, cancel, "")
				if err != nil {
					log.Println("Fehler beim Absenden der Nachricht: ", err)
				}
			}
		}
	}()
}

// NewConfig() parses the serverport
func NewConfig() Config {
	var cfg Config

	flag.IntVar(&cfg.port, "port", 8080, "HTTP Server Port")
	flag.Parse()

	return cfg
}
