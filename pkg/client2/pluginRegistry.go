package client2

type PluginInterface interface {
	// if an error accures, response.Content is empty
	Execute(message *Message) func() error
	Description() string
}

type PluginRegistry struct {
	plugins   map[string]PluginInterface
	invisible []string
}

// RegisterPlugins sets up all the plugins
func RegisterPlugins(chatService *ChatClient) *PluginRegistry {
	pr := PluginRegistry{plugins: make(map[string]PluginInterface)}
	pr.plugins["/help"] = NewHelpPlugin(chatService)
	pr.plugins["/time"] = NewTimePlugin(chatService)
	pr.plugins["/users"] = NewUserPlugin(chatService)
	pr.plugins["/register"] = NewRegisterClientPlugin(chatService)
	pr.plugins["/broadcast"] = NewBroadcastPlugin(chatService)
	pr.plugins["/quit"] = NewLogOutPlugin(chatService)
	pr.plugins["/private"] = NewPrivateMessagePlugin(chatService)

	pr.invisible = append(pr.invisible, "/broadcast")

	return &pr
}

func (pr *PluginRegistry) FindAndExecute(message *Message) func() error {
	plugin, ok := pr.plugins[message.Plugin]
	if !ok {
		return nil
	}

	return plugin.Execute(message)
}

// // ListPlugins lists all Plugins with correspontig commands
// func (pr *PluginRegistry) ListPlugins() []PluginInterface {
// 	plugins := []PluginInterface{}

// 	for command, plugin := range pr.plugins {
// 		if !slices.Contains(pr.invisible, command) {
// 			plugins = append(plugins, plugin)
// 		}
// 	}

// 	return plugins
// }
