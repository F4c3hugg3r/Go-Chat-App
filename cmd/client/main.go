package main

import (
	"flag"
	"fmt"
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

	// ctrlCBinding := prompt.KeyBind{
	// 	Key: prompt.ControlC,
	// 	Fn:  func(b *prompt.Buffer) { interChan <- os.Interrupt },
	// }

	// deleteInput := func(*prompt.Document) { fmt.Print("\033[1A\033[K") }

	signal.Notify(interChan, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	go interruptListener(interChan, c)

	if _, err := programm.Run(); err != nil {
		log.Fatal(err)
	}

	// p := prompt.New(
	// 	u.Executor,
	// 	u.Completer,
	// 	prompt.OptionAddKeyBind(ctrlCBinding),
	// 	prompt.OptionPrefix(""),
	// 	prompt.OptionBreakLineCallback(deleteInput),
	// )

	//TODO das beim starten des Outputs printen
	fmt.Println("-> registriere dich mit '/register {name}'")
	// p.Run()
}

// interruptListener sends a cancel() signal and closes all connections and requests if a interruption like
// os.Interrupt or syscall.SIGTERM is being triggered
func interruptListener(interChan chan os.Signal, c *n.ChatClient) {
	<-interChan

	if c.Registered {
		err := c.PostDelete(c.CreateMessage("", "/quit", "", ""))
		if err != nil {
			log.Print(err)
		}
	}

	c.HttpClient.CloseIdleConnections()

	log.Println("exiting programm")
	os.Exit(0)
}
