package client2

const (
	UnregisteredOnly = iota
	RegisteredOnly
	Always
)

type PluginInterface interface {
	// if an error accures, response.Content is empty
	Execute(message *Message) func() error
	Description() string
}

type PluginRegistry struct {
	plugins           map[string]PluginInterface
	registrationScope map[string]int
	invisible         []string
	chatClient        *ChatClient
}

// RegisterPlugins sets up all the plugins
func RegisterPlugins(chatClient *ChatClient) *PluginRegistry {
	pr := PluginRegistry{plugins: make(map[string]PluginInterface)}
	pr.plugins["/help"] = NewHelpPlugin(chatClient)
	pr.plugins["/time"] = NewTimePlugin(chatClient)
	pr.plugins["/users"] = NewUserPlugin(chatClient)
	pr.plugins["/register"] = NewRegisterClientPlugin(chatClient)
	pr.plugins["/broadcast"] = NewBroadcastPlugin(chatClient)
	pr.plugins["/quit"] = NewLogOutPlugin(chatClient)
	pr.plugins["/private"] = NewPrivateMessagePlugin(chatClient)

	pr.registrationScope = make(map[string]int)
	pr.registrationScope["/help"] = Always
	pr.registrationScope["/time"] = Always
	pr.registrationScope["/register"] = UnregisteredOnly
	pr.registrationScope["/users"] = RegisteredOnly
	pr.registrationScope["/broadcast"] = RegisteredOnly
	pr.registrationScope["/quit"] = RegisteredOnly
	pr.registrationScope["/private"] = RegisteredOnly

	pr.invisible = append(pr.invisible, "/broadcast")
	pr.chatClient = chatClient

	return &pr
}

func (pr *PluginRegistry) FindAndExecute(message *Message) func() error {
	plugin, ok := pr.plugins[message.Plugin]
	if !ok {
		return nil
	}

	scope, ok := pr.registrationScope[message.Plugin]
	if !ok {
		return nil
	}

	switch scope {
	case UnregisteredOnly:
		if pr.chatClient.Registered {
			return nil
		}
	case RegisteredOnly:
		if !pr.chatClient.Registered {
			return nil
		}
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
