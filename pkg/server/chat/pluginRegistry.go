package chat

import (
	"fmt"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/server/types"
)

type PluginInterface interface {
	Execute(message *ty.Message) (*ty.Response, error)
	Description() *Description
}

type PluginRegistry struct {
	plugins map[string]PluginInterface
}

type Plugin struct {
	Command     string
	Description string
}

type Description struct {
	Description string
	Template    string
}

// RegisterPlugins sets up all the plugins
func RegisterPlugins(chatService *ChatService) *PluginRegistry {
	pr := &PluginRegistry{plugins: make(map[string]PluginInterface)}
	pr.plugins["/help"] = NewHelpPlugin(pr)
	pr.plugins["/time"] = NewTimePlugin()
	pr.plugins["/users"] = NewListUsersPlugin(chatService)
	pr.plugins["/register"] = NewRegisterClientPlugin(chatService, pr)
	pr.plugins["/broadcast"] = NewBroadcastPlugin(chatService)
	pr.plugins["/quit"] = NewLogOutPlugin(chatService, pr)
	pr.plugins["/private"] = NewPrivateMessagePlugin(chatService)
	pr.plugins["/group"] = RegisterGroupPlugins(chatService, pr)

	return pr
}

func (pr *PluginRegistry) FindAndExecute(message *ty.Message) (*ty.Response, error) {
	plugin, ok := pr.plugins[message.Plugin]
	if !ok {
		return &ty.Response{Err: fmt.Sprintf("%v: no such chat plugin found: %s", ty.ErrNoPermission, message.Plugin)}, nil
	}

	return plugin.Execute(message)
}
