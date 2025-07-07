package client2

import (
	"fmt"
	"strings"
)

// PrivateMessage Plugin lets a client send a private message to another client identified by it's clientId
type PrivateMessagePlugin struct {
	c *ChatClient
}

func NewPrivateMessagePlugin(chatClient *ChatClient) *PrivateMessagePlugin {
	return &PrivateMessagePlugin{c: chatClient}
}

func (pp *PrivateMessagePlugin) Description() string {
	return "lets you send a private message to someone \n-> template: '/private {Id} {message}'"
}

func (pp *PrivateMessagePlugin) Execute(message *Message) func() error {
	return func() error {
		opposingClientId := strings.Fields(message.Content)[1]

		content, ok := strings.CutPrefix(message.Content, fmt.Sprintf("/private %s ", opposingClientId))
		if !ok {
			return fmt.Errorf("%w: prefix '/private %s ' not found", ErrParsing, opposingClientId)
		}

		return pp.c.SendPlugin(pp.c.CreateMessage(message.Name, message.Plugin, content, opposingClientId))
	}
}

// LogOutPlugin logs out a client by deleting it out of the clients map
type LogOutPlugin struct {
	c *ChatClient
}

func NewLogOutPlugin(chatClient *ChatClient) *LogOutPlugin {
	return &LogOutPlugin{c: chatClient}
}

func (lp *LogOutPlugin) Description() string {
	return "loggs you out of the chat"
}

func (lp *LogOutPlugin) Execute(message *Message) func() error {
	return func() error {
		return lp.c.SendDelete(message)
	}
}

// RegisterClientPlugin safely registeres a client by creating a Client with the received values
// and putting it into the global clients map
type RegisterClientPlugin struct {
	c *ChatClient
}

func NewRegisterClientPlugin(chatClient *ChatClient) *RegisterClientPlugin {
	return &RegisterClientPlugin{c: chatClient}
}

func (rp *RegisterClientPlugin) Description() string {
	return "registeres a client"
}

func (rp *RegisterClientPlugin) Execute(message *Message) func() error {
	return func() error {

		clientName, ok := strings.CutPrefix(message.Content, "/register ")
		if !ok {
			return fmt.Errorf("%w: prefix '/register ' not found", ErrParsing)
		}

		if len(clientName) > 50 {
			return fmt.Errorf("%w: name %s is too long", ErrParsing, clientName)
		}

		return rp.c.SendRegister(rp.c.CreateMessage(clientName, message.Plugin, clientName, message.ClientId))
	}
}

// BroadcaastPlugin distributes an incomming message abroad all client channels if
// a client can't receive, i'ts active status is set to false
type BroadcastPlugin struct {
	c *ChatClient
}

func NewBroadcastPlugin(chatClient *ChatClient) *BroadcastPlugin {
	return &BroadcastPlugin{c: chatClient}
}

func (bp *BroadcastPlugin) Description() string {
	return "distributes a message abroad all clients"
}

func (bp *BroadcastPlugin) Execute(message *Message) func() error {
	return func() error {
		return bp.c.SendPlugin(message)
	}
}

// HelpPlugin tells you information about available plugins
type HelpPlugin struct {
	c *ChatClient
}

func NewHelpPlugin(chatClient *ChatClient) *HelpPlugin {
	return &HelpPlugin{c: chatClient}
}

func (h *HelpPlugin) Description() string {
	return "tells every plugin and their description"
}

func (h *HelpPlugin) Execute(message *Message) func() error {
	return func() error {
		return h.c.SendPlugin(message)
	}
}

// UserPlugin tells you information about all the current users
type UserPlugin struct {
	c *ChatClient
}

func NewUserPlugin(chatClient *ChatClient) *UserPlugin {
	return &UserPlugin{c: chatClient}
}

func (u *UserPlugin) Description() string {
	return "tells you information about all the current users"
}

func (u *UserPlugin) Execute(message *Message) func() error {
	return func() error {
		return u.c.SendPlugin(message)
	}
}

// TimePlugin tells you the current time
type TimePlugin struct {
	c *ChatClient
}

func NewTimePlugin(chatClient *ChatClient) *TimePlugin {
	return &TimePlugin{c: chatClient}
}

func (t *TimePlugin) Description() string {
	return "tells you the current time"
}

func (t *TimePlugin) Execute(message *Message) func() error {
	return func() error {
		return t.c.SendPlugin(message)
	}
}
