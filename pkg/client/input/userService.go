package input

import (
	"fmt"
	"strings"
	"sync"

	n "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/network"
	p "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/plugins"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/types"
	"github.com/c-bata/go-prompt"
)

// UserService handles user inputs and outputs
type UserService struct {
	ChatClient *n.ChatClient
	plugins    *p.PluginRegistry
	poll       bool
	typing     bool
	mu         *sync.RWMutex
	cond       *sync.Cond
}

// NewUserService creates a UserService
func NewUserService(c *n.ChatClient) *UserService {
	u := &UserService{
		ChatClient: c,
		plugins:    p.RegisterPlugins(c),
		poll:       false,
		mu:         &sync.RWMutex{},
	}

	u.cond = sync.NewCond(u.mu)

	// go u.ResponsePoller()

	return u
}

// ResponsePoller gets and displays messages if the client is not typing
func (u *UserService) ResponsePoller() *t.Response {
	var rsp *t.Response

	// for {
	// u.checkPolling()

	// select {
	// case rsp = <-u.chatClient.Output:
	rsp, ok := <-u.ChatClient.Output
	if !ok {
		return &t.Response{Err: fmt.Errorf("%w: channel is closed", t.ErrNoPermission)}
	}
	// err := u.DisplayResponse(rsp)
	// if err != nil {

	// 	// TODO in output channel pushen
	// 	log.Printf("%v: response from %s couldn't be displayed", err, rsp.Name)
	// }
	// default:
	// 	// bei <-time.After() würde es zu potenziellen synchronisations-Problemen kommen, wenn das polling gestoppt wird
	// 	time.Sleep(100 * time.Millisecond)
	// // }
	// }
	return rsp
}

// stopPoll stopps the polling
// func (u *UserService) stopPoll() {
// 	u.mu.Lock()
// 	defer u.mu.Unlock()

// 	u.poll = false
// }

// // startPoll starts the polling
// func (u *UserService) startPoll() {
// 	u.mu.Lock()
// 	defer u.mu.Unlock()

// 	u.poll = true
// 	u.Cond.Signal()
// }

// // checkPolling blocks until polling is started
// func (u *UserService) checkPolling() {
// 	u.mu.Lock()
// 	defer u.mu.Unlock()

// 	for !u.poll {
// 		u.Cond.Wait()
// 	}
// }

// ParseInputToMessage parses the user input into a Message
func (u *UserService) ParseInputToMessage(input string) *t.Message {
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

	return u.ChatClient.CreateMessage("", plugin, content, "")
}

// Executor takes the parsed input message, executes the corresponding plugin
func (u *UserService) Executor(input string) {
	msg := u.ParseInputToMessage(input)

	err, comment := u.plugins.FindAndExecute(msg)
	if err != nil {
		u.ChatClient.Output <- &t.Response{Err: fmt.Errorf("%v: %s", err.Error(), err)}
	}

	u.ChatClient.Output <- &t.Response{Err: fmt.Errorf("%s", comment)}
}

// // IsTyping receives the length of the userinput, checks if the client
// // is typing and sets the typing parameter
// func (u *UserService) isTyping(sliceLength int) bool {
// 	switch sliceLength {
// 	case 0:
// 		u.typing = false
// 		return false
// 	default:
// 		u.typing = true
// 		return true
// 	}
// }

// Completer suggests plugins and their descriptions in the stdIn
func (u *UserService) Completer(d prompt.Document) []prompt.Suggest {
	// u.stopPoll()

	s := []prompt.Suggest{}
	textBeforeCursor := d.TextBeforeCursor()
	words := strings.Fields(textBeforeCursor)

	// if !u.isTyping(len(words)) {
	// 	// u.startPoll()
	// }

	if len(words) == 1 && d.GetWordBeforeCursor() != "" {
		for command, plugin := range u.plugins.Plugins {
			s = append(s, prompt.Suggest{Text: command, Description: plugin.Description()})
		}

		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}

	return s
}
