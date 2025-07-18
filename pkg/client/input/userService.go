package input

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	n "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/network"
	p "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/plugins"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/types"
)

// UserService handles user inputs and outputs
type UserService struct {
	ChatClient *n.ChatClient
	PlugReg    *p.PluginRegistry
	poll       bool
	typing     bool
	mu         *sync.RWMutex
	cond       *sync.Cond
}

// NewUserService creates a UserService
func NewUserService(c *n.ChatClient) *UserService {
	u := &UserService{
		ChatClient: c,
		PlugReg:    p.RegisterPlugins(c),
		poll:       false,
		mu:         &sync.RWMutex{},
	}

	u.cond = sync.NewCond(u.mu)

	return u
}

func (u *UserService) HandleAddGroup(groupJson string) (*t.Group, error) {
	group, err := decodeStringToGroup(groupJson)
	if err != nil {
		return nil, err
	}

	u.ChatClient.SetGroupId(group.GroupId)
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

	rsp, ok := <-u.ChatClient.Output
	if !ok {
		return &t.Response{Err: fmt.Errorf("%w: channel is closed", t.ErrNoPermission)}
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

	return u.ChatClient.CreateMessage("", plugin, content, "")
}

// Executor takes the parsed input message, executes the corresponding plugin
func (u *UserService) Executor(input string) {
	msg := u.ParseInputToMessage(input)

	err, comment := u.PlugReg.FindAndExecute(msg)
	if err != nil {
		u.ChatClient.Output <- &t.Response{Err: fmt.Errorf("%v", err)}
		return
	}

	u.ChatClient.Output <- &t.Response{Content: comment}
}

func decodeStringToGroup(jsonGroup string) (*t.Group, error) {
	var group *t.Group
	dec := json.NewDecoder(strings.NewReader(jsonGroup))
	err := dec.Decode(&group)
	return group, err
}
