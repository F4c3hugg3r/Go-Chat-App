package client2

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/c-bata/go-prompt"
)

// NewUserService creates a UserService
func NewUserService(c *ChatClient) *UserService {
	u := &UserService{
		chatClient: c,
		plugins:    RegisterPlugins(c),
		poll:       false,
		mu:         &sync.Mutex{},
	}

	u.Cond = sync.NewCond(u.mu)

	go u.ResponsePoller()

	return u
}

// ResponsePoller gets and displays messages if the client is not typing
func (u *UserService) ResponsePoller() {
	var rsp *Response
	for {
		u.CheckPolling()

		select {
		case rsp = <-u.chatClient.Output:
			err := u.DisplayResponse(rsp)
			if err != nil {
				log.Printf("%v: response from %s couldn't be displayed", err, rsp.Name)
			}
		default:
			// bei <-time.After() würde es zu potenziellen synchronisations-Problemen
			// kommen, wenn das polling gestoppt wird
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// stopPoll stopps the polling
func (u *UserService) stopPoll() {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.poll = false
}

// startPoll starts the polling
func (u *UserService) startPoll() {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.poll = true
	u.Cond.Signal()
}

// CheckPolling blocks until polling is started
func (u *UserService) CheckPolling() {
	u.mu.Lock()
	defer u.mu.Unlock()

	for !u.poll {
		u.Cond.Wait()
	}
}

// DisplayResponse prints out a Response in the proper way
func (u *UserService) DisplayResponse(rsp *Response) error {
	if rsp.Content == "" {
		return nil
	}

	if strings.HasPrefix(rsp.Content, "[") {
		output, err := JSONToTable(rsp.Content)
		if err != nil {
			return fmt.Errorf("%v: error formatting json to table", err)
		}

		fmt.Println(output)
		return nil
	}

	responseString := fmt.Sprintf("%s: %s", rsp.Name, rsp.Content)
	fmt.Println(responseString)
	return nil
}

// ParseInputToMessage parses the user input into a Message
func (u *UserService) ParseInputToMessage(input string) (*Message, error) {
	input = strings.TrimSuffix(input, "\n")
	var plugin string

	ok := strings.HasPrefix(input, "/")
	switch ok {
	case true:
		plugin = strings.Fields(input)[0]
	case false:
		plugin = "/broadcast"
	}

	content := strings.ReplaceAll(input, plugin, "")
	content, _ = strings.CutPrefix(content, " ")

	return u.chatClient.CreateMessage("", plugin, content, ""), nil
}

// Executor takes the parsed input message, executes the corresponding plugin
func (u *UserService) Executor(input string) {
	msg, err := u.ParseInputToMessage(input)
	if err != nil {
		log.Printf("%v: wrong input", err)
	}

	err = u.plugins.FindAndExecute(msg)()
	if err != nil {
		log.Printf("%v: couldn't send message", err)
	}
}

// IsTyping receives the length of the userinput, checks if the client
// is typing and sets the typing parameter
func (u *UserService) isTyping(sliceLength int) bool {
	switch sliceLength {
	case 0:
		u.typing = false
		return false
	default:
		u.typing = true
		return true
	}
}

// Completer suggests plugins and their descriptions in the stdIn
func (u *UserService) Completer(d prompt.Document) []prompt.Suggest {
	u.stopPoll()

	s := []prompt.Suggest{}
	textBeforeCursor := d.TextBeforeCursor()
	words := strings.Fields(textBeforeCursor)

	if !u.isTyping(len(words)) {
		u.startPoll()
	}

	if len(words) == 1 && d.GetWordBeforeCursor() != "" {
		for command, plugin := range u.plugins.plugins {
			s = append(s, prompt.Suggest{Text: command, Description: plugin.Description()})
		}
		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)

	}
	return s
}
