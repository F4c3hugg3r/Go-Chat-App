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

type UserPlugin struct {
	chatService *ChatService
}

func NewUserPlugin(s *ChatService) *UserPlugin {
	return &UserPlugin{chatService: s}
}

func (u *UserPlugin) Execute() []string {
	return u.chatService.ListClients()
}

type TimePlugin struct{}

func NewTimePlugin() *TimePlugin {
	return &TimePlugin{}
}

func (t *TimePlugin) Execute() []string {
	return []string{time.Now().String()}
}
