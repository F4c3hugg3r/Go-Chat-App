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

var (
	url = flag.String("url", "http://localhost:8080", "HTTP Server URL")
)

func init() {
	flag.Parse()
}

func main() {
	c := client.NewClient(*url)
	u := client.NewUserService(c)
	interChan := make(chan os.Signal, 3)

	ctrlCBinding := prompt.KeyBind{
		Key: prompt.ControlC,
		Fn:  func(b *prompt.Buffer) { interChan <- os.Interrupt },
	}

	signal.Notify(interChan, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go interruptListener(interChan, c)

	p := prompt.New(
		u.Executor,
		u.Completer,
		prompt.OptionAddKeyBind(ctrlCBinding),
		prompt.OptionPrefix(">> "),
	)
	p.Run()
}

// interruptListener sends a cancel() signal and closes all connections and requests if a interruption like
// os.Interrupt or syscall.SIGTERM is being triggered
func interruptListener(interChan chan os.Signal, c *client.ChatClient) {
	<-interChan

	err := c.SendDelete(c.CreateMessage("", "/quit", "", ""))
	if err != nil {
		log.Print(err)
	}

	c.HttpClient.CloseIdleConnections()

	log.Println("Client logged out")
	os.Exit(0)
}
