package plugins

import (
	"fmt"
	"strings"

	n "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/network"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/types"
)

// PrivateMessage Plugin lets a client send a private message to another client identified by it's clientId
type PrivateMessagePlugin struct {
	c *n.ChatClient
}

func NewPrivateMessagePlugin(chatClient *n.ChatClient) *PrivateMessagePlugin {
	return &PrivateMessagePlugin{c: chatClient}
}

func (pp *PrivateMessagePlugin) CheckScope() int {
	return RegisteredOnly
}

func (pp *PrivateMessagePlugin) Description() string {
	return "lets you send a private message to someone"
}

func (pp *PrivateMessagePlugin) Execute(message *t.Message) (error, string) {
	opposingClientId := strings.Fields(message.Content)[0]

	content, ok := strings.CutPrefix(message.Content, fmt.Sprintf("%s ", opposingClientId))
	if !ok {
		return fmt.Errorf("%w: prefix '%s ' not found", t.ErrParsing, opposingClientId), ""
	}

	_, err := pp.c.PostMessage(pp.c.CreateMessage(message.Name, message.Plugin, content, opposingClientId), t.PostPlugin)

	return err, ""
}

// LogOutPlugin logs out a client by deleting it out of the clients map
type LogOutPlugin struct {
	c *n.ChatClient
}

func NewLogOutPlugin(chatClient *n.ChatClient) *LogOutPlugin {
	return &LogOutPlugin{c: chatClient}
}

func (lp *LogOutPlugin) CheckScope() int {
	return RegisteredOnly
}

func (lp *LogOutPlugin) Description() string {
	return "loggs you out of the chat"
}

func (lp *LogOutPlugin) Execute(message *t.Message) (error, string) {
	return lp.c.PostDelete(message), "- Du bist nun vom Server getrennt -"
}

// RegisterClientPlugin safely registeres a client by creating a Client with the received values
// and putting it into the global clients map
type RegisterClientPlugin struct {
	c *n.ChatClient
}

func NewRegisterClientPlugin(chatClient *n.ChatClient) *RegisterClientPlugin {
	return &RegisterClientPlugin{c: chatClient}
}

func (rp *RegisterClientPlugin) CheckScope() int {
	return UnregisteredOnly
}

func (rp *RegisterClientPlugin) Description() string {
	return "registeres a client"
}

func (rp *RegisterClientPlugin) Execute(message *t.Message) (error, string) {
	clientName := message.Content
	if len(clientName) > 50 || len(clientName) < 3 {
		return fmt.Errorf("%w: your name has to be between 3 and 50 chars long", t.ErrParsing), ""
	}

	rsp, err := rp.c.PostMessage(rp.c.CreateMessage(clientName, message.Plugin, message.Content, message.ClientId), t.PostRegister)
	if err != nil {
		return fmt.Errorf("%w: error sending message", err), ""
	}

	err = rp.c.Register(rsp)
	if err != nil {
		return fmt.Errorf("%w: error registering client", err), ""
	}

	return err, "- Du wurdest registriert -\n-> Gebe '/quit' ein, um den Chat zu verlassen\n-> Oder '/help' um Commands auzuführen\n-> Oder ctrl+C / Esc um das Programm zu schließen"
}

// BroadcaastPlugin distributes an incomming message abroad all client channels if
// a client can't receive, i'ts active status is set to false
type BroadcastPlugin struct {
	c *n.ChatClient
}

func NewBroadcastPlugin(chatClient *n.ChatClient) *BroadcastPlugin {
	return &BroadcastPlugin{c: chatClient}
}

func (bp *BroadcastPlugin) CheckScope() int {
	return RegisteredOnly
}

func (bp *BroadcastPlugin) Description() string {
	return "distributes a message abroad all clients"
}

func (bp *BroadcastPlugin) Execute(message *t.Message) (error, string) {
	_, err := bp.c.PostMessage(message, t.PostPlugin)
	return err, ""
}

// HelpPlugin tells you information about available plugins
type HelpPlugin struct {
	c *n.ChatClient
}

func NewHelpPlugin(chatClient *n.ChatClient) *HelpPlugin {
	return &HelpPlugin{c: chatClient}
}

func (hp *HelpPlugin) CheckScope() int {
	return RegisteredOnly
}

func (h *HelpPlugin) Description() string {
	return "tells every plugin and their description"
}

func (h *HelpPlugin) Execute(message *t.Message) (error, string) {
	_, err := h.c.PostMessage(message, t.PostPlugin)
	return err, ""
}

// UserPlugin tells you information about all the current users
type UserPlugin struct {
	c *n.ChatClient
}

func NewUserPlugin(chatClient *n.ChatClient) *UserPlugin {
	return &UserPlugin{c: chatClient}
}

func (up *UserPlugin) CheckScope() int {
	return RegisteredOnly
}

func (u *UserPlugin) Description() string {
	return "tells you information about all the current users"
}

func (u *UserPlugin) Execute(message *t.Message) (error, string) {
	_, err := u.c.PostMessage(message, t.PostPlugin)
	return err, ""
}

// TimePlugin tells you the current time
type TimePlugin struct {
	c *n.ChatClient
}

func NewTimePlugin(chatClient *n.ChatClient) *TimePlugin {
	return &TimePlugin{c: chatClient}
}

func (tp *TimePlugin) CheckScope() int {
	return RegisteredOnly
}

func (tp *TimePlugin) Description() string {
	return "tells you the current time"
}

func (tp *TimePlugin) Execute(message *t.Message) (error, string) {
	_, err := tp.c.PostMessage(message, t.PostPlugin)
	return err, ""
}
