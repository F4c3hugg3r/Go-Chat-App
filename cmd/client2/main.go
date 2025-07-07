package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	client "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client2"
	"github.com/c-bata/go-prompt"
)

// globale var für cfg
type Config struct {
	url string
}

func main() {
	cfg := NewConfig()
	c := client.NewClient(cfg.url)
	u := client.NewUserService(c)

	interChan := make(chan os.Signal, 2)
	signal.Notify(interChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// doesn't interrupt properly nimmt wahrscheinlich ^C nicht richtig
	go interruptListener(interChan, c, cfg.url)

	p := prompt.New(
		u.Executor,
		u.Completer,
		//maybe schauen ob es was anderes praktisches gibt
		prompt.OptionPrefix(">> "),
	)
	p.Run()
}

// interruptListener sends a cancel() signal and closes all connections and requests if a interruption like
// os.Interrupt or syscall.SIGTERM is being triggered
func interruptListener(interChan chan os.Signal, client *client.ChatClient, url string) {
	<-interChan
	//TODO sendDelete
	// err := client.SendDelete(nil)
	// if err != nil {
	// 	log.Print(err)
	// }

	client.HttpClient.CloseIdleConnections()

	log.Println("Client logged out")
	os.Exit(0)
}

// NewConfig() parses the serverport
func NewConfig() Config {
	var cfg Config

	//url statt port
	// in init fnc
	flag.StringVar(&cfg.url, "url", "http://localhost:8080", "HTTP Server URL")

	flag.Parse()

	return cfg
}
