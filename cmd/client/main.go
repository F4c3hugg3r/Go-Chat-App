package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	ui "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/UI"
	i "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/input"
	n "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/network"
)

var (
	url = flag.String("url", "http://localhost:8080", "HTTP Server URL")
)

func init() {
	flag.Parse()
}

func main() {
	c := n.NewClient(*url)
	u := i.NewUserService(c)
	programm := tea.NewProgram(ui.InitialModel(u))
	interChan := make(chan os.Signal, 3)

	signal.Notify(interChan, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt)

	go interruptListener(interChan, c)

	if _, err := programm.Run(); err != nil {
		log.Fatal(err)
	}
}

// interruptListener sends a cancel() signal and closes all connections and requests if a interruption like
// os.Interrupt or syscall.SIGTERM is being triggered
func interruptListener(interChan chan os.Signal, c *n.Client) {
	<-interChan

	if c.Registered {
		c.Interrupt()
	}

	c.HttpClient.CloseIdleConnections()

	log.Println("exiting programm")
	os.Exit(0)
}
