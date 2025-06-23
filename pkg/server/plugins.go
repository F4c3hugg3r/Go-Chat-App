package server

import (
	"fmt"
	"time"
)

type PluginInterface interface {
	Execute() []string
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

func (pr *PluginRegistry) FindAndExecute(command string) (result []string, err error) {
	if plugin, ok := pr.plugins[command]; ok {
		return plugin.Execute(), nil
	} else {
		return []string{}, fmt.Errorf("no such plugin found: %s", command)
	}
}

// TimePlugin tells you the current time
type HelpPlugin struct {
	pr *PluginRegistry
}

func NewHelpPlugin(pr *PluginRegistry) *HelpPlugin {
	return &HelpPlugin{pr: pr}
}

func (h *HelpPlugin) Execute() []string {
	return ListPlugins(h.pr)
}

// ListPlugins lists all Plugins with correspontig commands
func ListPlugins(pr *PluginRegistry) []string {
	slice := []string{"You can execute a command by typing one of the the following commands:"}

	for commands := range pr.plugins {
		slice = append(slice, commands)
	}
	return slice
}

// UserPlugin tells you information about all the current users
type UserPlugin struct {
	chatService *ChatService
}

func NewUserPlugin(s *ChatService) *UserPlugin {
	return &UserPlugin{chatService: s}
}

func (u *UserPlugin) Execute() []string {
	return u.chatService.ListClients()
}

// TimePlugin tells you the current time
type TimePlugin struct{}

func NewTimePlugin() *TimePlugin {
	return &TimePlugin{}
}

func (t *TimePlugin) Execute() []string {
	return []string{time.Now().String()}
}
