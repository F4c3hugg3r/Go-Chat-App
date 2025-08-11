package input

import (
	"fmt"
	"strings"
	"sync"

	n "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/network"
	p "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/plugins"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// UserService handles user inputs and outputs
type UserService struct {
	Client  *n.Client
	PlugReg *p.PluginRegistry
	poll    bool
	typing  bool
	mu      *sync.RWMutex
	cond    *sync.Cond
	// logging
	LoggChan chan t.Log
}

// NewUserService creates a UserService
func NewUserService(c *n.Client) *UserService {
	u := &UserService{
		Client:  c,
		poll:    false,
		mu:      &sync.RWMutex{},
		PlugReg: p.RegisterPlugins(c),
	}

	u.cond = sync.NewCond(u.mu)

	return u
}

func (u *UserService) HandleAddGroup(groupJson string) (*t.JsonGroup, error) {
	group, err := t.DecodeStringToJsonGroup(groupJson)
	if err != nil {
		return nil, err
	}

	u.Client.SetGroupId(group.GroupId)
	return group, nil
}

func (u *UserService) InitializeSuggestions() []string {
	s := []string{}

	for command := range u.PlugReg.Plugins {
		s = append(s, command)
	}

	return s
}

// ResponsePoller gets and displays messages if the client is not typing
func (u *UserService) ResponsePoller() *t.Response {
	var rsp *t.Response

	rsp, ok := <-u.Client.Output
	if !ok {
		return &t.Response{Err: fmt.Sprintf("%v: channel is closed", t.ErrNoPermission)}
	}

	return rsp
}

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

	return u.Client.CreateMessage("", plugin, content, "")
}

// Executor takes the parsed input message, executes the corresponding plugin
func (u *UserService) Executor(input string) {
	msg := u.ParseInputToMessage(input)

	err, comment := u.PlugReg.FindAndExecute(msg)
	if err != nil {
		u.Client.Output <- &t.Response{Err: err.Error()}
		return
	}

	u.Client.Output <- &t.Response{Content: comment}
}
