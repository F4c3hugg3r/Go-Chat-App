package client2

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	client "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client2"
	"github.com/c-bata/go-prompt"
)

type Config struct {
	url string
}

func main() {
	cfg := NewConfig()
	url := fmt.Sprintf("http://localhost:%d", cfg.url)
	c := client.NewClient(url)

	interChan := make(chan os.Signal, 2)
	signal.Notify(interChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go interruptListener(interChan, c, url)

	p := prompt.New(
		c.Executor,
		c.Completer,
		//maybe schauen ob es was anderes praktisches gibt
		prompt.OptionPrefix(">> "),
	)
	p.Run()
}

// interruptListener sends a cancel() signal and closes all connections and requests if a interruption like
// os.Interrupt or syscall.SIGTERM is being triggered
func interruptListener(interChan chan os.Signal, client *client.ChatClient, url string) {
	<-interChan
	//err := client.SendMessage(url, cancel, "/quit\n", wg, ctx)
	// if err != nil {
	// 	log.Print(err)
	// }

	client.HttpClient.CloseIdleConnections()

	log.Println("Client logged out")
	os.Exit(69)
}

// NewConfig() parses the serverport
func NewConfig() Config {
	var cfg Config

	//url statt port
	flag.StringVar(&cfg.url, "url", "http://localhost:8080", "HTTP Server URL")
	flag.Parse()

	return cfg
}
