package server

import (
	"fmt"
	"time"
)

type PluginInterface interface {
	//consumes Message and produces Response
	Execute(message *Message) (Response, error)
}

type PluginRegistry struct {
	plugins map[string]PluginInterface
}

func RegisterPlugins(chatService *ChatService) *PluginRegistry {
	pr := PluginRegistry{plugins: make(map[string]PluginInterface)}
	pr.plugins["/help"] = NewHelpPlugin(&pr)
	pr.plugins["/time"] = NewTimePlugin()
	pr.plugins["/users"] = NewUserPlugin(chatService)
	return &pr
}

func (pr *PluginRegistry) FindAndExecute(message *Message) (Response, error) {
	if plugin, ok := pr.plugins[message.Plugin]; ok {
		return plugin.Execute(message)
	}
	return Response{message.Name, fmt.Sprintf("no such plugin found: %s", message.Plugin)}, fmt.Errorf("no such plugin found: %s", message.Plugin)
}

// TimePlugin tells you the current time
type HelpPlugin struct {
	pr *PluginRegistry
}

func NewHelpPlugin(pr *PluginRegistry) *HelpPlugin {
	return &HelpPlugin{pr: pr}
}

func (h *HelpPlugin) Execute(message *Message) (Response, error) {
	ListPlugins(h.pr)
	return Response{message.Plugin, "TODO"}, nil
}

// ListPlugins lists all Plugins with correspontig commands
func ListPlugins(pr *PluginRegistry) string {
	slice := []string{"You can execute a command by typing one of the the following commands:"}

	for commands := range pr.plugins {
		slice = append(slice, commands)
	}
	return "TODO"
}

// UserPlugin tells you information about all the current users
type UserPlugin struct {
	chatService *ChatService
}

func NewUserPlugin(s *ChatService) *UserPlugin {
	return &UserPlugin{chatService: s}
}

func (u *UserPlugin) Execute(messsage *Message) (Response, error) {
	u.chatService.ListClients()
	return Response{messsage.Name, "TODO"}, nil
}

// TimePlugin tells you the current time
type TimePlugin struct{}

func NewTimePlugin() *TimePlugin {
	return &TimePlugin{}
}

func (t *TimePlugin) Execute(message *Message) (Response, error) {
	return Response{message.Name, time.Now().String()}, nil
}
