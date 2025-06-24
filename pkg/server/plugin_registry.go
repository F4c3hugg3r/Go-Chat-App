package server

import (
	"fmt"
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
