package plugins

import (
	"fmt"

	n "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/network"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

const (
	UnregisteredOnly = iota
	RegisteredOnly
	// InGroupOnly
	Always
)

// PluginInterface describes the plugin methods
type PluginInterface interface {
	Execute(message *t.Message) (error, string)
	CheckScope() int
}

// PluginRegistry contains the plugins and their methods
type PluginRegistry struct {
	Plugins    map[string]PluginInterface
	chatClient *n.ChatClient
}

// logging
// RegisterPlugins sets up all the plugins
func RegisterPlugins(chatClient *n.ChatClient, logChan chan t.Logg) *PluginRegistry {
	pr := PluginRegistry{Plugins: make(map[string]PluginInterface)}
	pr.Plugins["/help"] = NewHelpPlugin(chatClient)
	pr.Plugins["/time"] = NewTimePlugin(chatClient)
	pr.Plugins["/users"] = NewUserPlugin(chatClient)
	pr.Plugins["/register"] = NewRegisterClientPlugin(chatClient)
	pr.Plugins["/broadcast"] = NewBroadcastPlugin(chatClient)
	pr.Plugins["/quit"] = NewLogOutPlugin(chatClient)
	pr.Plugins["/private"] = NewPrivateMessagePlugin(chatClient)
	pr.Plugins["/group"] = NewGroupPlugin(chatClient)
	pr.Plugins["/call"] = NewCallPlugin(chatClient, logChan)

	pr.chatClient = chatClient

	return &pr
}

// FindAndExecute executes the plugins Execute method if the scope is fitting
func (pr *PluginRegistry) FindAndExecute(message *t.Message) (error, string) {
	plugin, ok := pr.Plugins[message.Plugin]
	if !ok {
		return fmt.Errorf("%w: plugin not found", t.ErrNoPermission), ""
	}

	scope := pr.Plugins[message.Plugin].CheckScope()

	switch scope {
	case UnregisteredOnly:
		if pr.chatClient.Registered {
			return fmt.Errorf("%w: you are already registered", t.ErrNoPermission), ""
		}
	case RegisteredOnly:
		if !pr.chatClient.Registered {
			return fmt.Errorf("%w: you are not registered yet", t.ErrNoPermission), ""
		}
		// case InGroupOnly:
		// 	if pr.chatClient.GetGroupId() == "" {
		// 		return fmt.Errorf("%w: you are not in a group yet", t.ErrNoPermission), ""
		// 	}
	}

	return plugin.Execute(message)
}
