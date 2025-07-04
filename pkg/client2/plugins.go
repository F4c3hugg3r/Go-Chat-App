package client2

// // PrivateMessage Plugin lets a client send a private message to another client identified by it's clientId
// type PrivateMessagePlugin struct {
// 	client *ChatClient
// }

// func NewPrivateMessagePlugin(s *ChatClient) *PrivateMessagePlugin {
// 	return &PrivateMessagePlugin{client: s}
// }

// func (pp *PrivateMessagePlugin) Description() string {
// 	return "lets you send a private message to someone \n-> template: '/private {Id} {message}'"
// }

// func (pp *PrivateMessagePlugin) Execute(message *Message) error {
// 	return nil
// }

// // LogOutPlugin logs out a client by deleting it out of the clients map
// type LogOutPlugin struct {
// 	client *ChatClient
// }

// func NewLogOutPlugin(s *ChatClient) *LogOutPlugin {
// 	return &LogOutPlugin{client: s}
// }

// func (lp *LogOutPlugin) Description() string {
// 	return "loggs you out of the chat"
// }

// func (lp *LogOutPlugin) Execute(message *Message) error {
// 	return nil
// }

// // RegisterClientPlugin safely registeres a client by creating a Client with the received values
// // and putting it into the global clients map
// type RegisterClientPlugin struct {
// 	client *ChatClient
// }

// func NewRegisterClientPlugin(s *ChatClient) *RegisterClientPlugin {
// 	return &RegisterClientPlugin{client: s}
// }

// func (rp *RegisterClientPlugin) Description() string {
// 	return "registeres a client"
// }

// func (rp *RegisterClientPlugin) Execute(message *Message) error {
// 	return nil
// }

// // BroadcaastPlugin distributes an incomming message abroad all client channels if
// // a client can't receive, i'ts active status is set to false
// type BroadcastPlugin struct {
// 	client *ChatClient
// }

// func NewBroadcastPlugin(s *ChatClient) *BroadcastPlugin {
// 	return &BroadcastPlugin{client: s}
// }

// func (bp *BroadcastPlugin) Description() string {
// 	return "distributes a message abroad all clients"
// }

// func (bp *BroadcastPlugin) Execute(message *Message) error {
// 	return nil
// }

// // HelpPlugin tells you information about available plugins
// type HelpPlugin struct {
// 	pr *PluginRegistry
// }

// func NewHelpPlugin(pr *PluginRegistry) *HelpPlugin {
// 	return &HelpPlugin{pr: pr}
// }

// func (h *HelpPlugin) Description() string {
// 	return "tells every plugin and their description"
// }

// func (h *HelpPlugin) Execute(message *Message) error {
// 	return nil
// }

// // UserPlugin tells you information about all the current users
// type UserPlugin struct {
// 	client *ChatClient
// }

// func NewUserPlugin(s *ChatClient) *UserPlugin {
// 	return &UserPlugin{client: s}
// }

// func (u *UserPlugin) Description() string {
// 	return "tells you information about all the current users"
// }

// func (u *UserPlugin) Execute(message *Message) error {
// 	return nil
// }

// // TimePlugin tells you the current time
// type TimePlugin struct{}

// func NewTimePlugin() *TimePlugin {
// 	return &TimePlugin{}
// }

// func (t *TimePlugin) Description() string {
// 	return "tells you the current time"
// }

// func (t *TimePlugin) Execute(message *Message) error {
// 	body, err := json.Marshal(&message)
// 	if err != nil {
// 		return fmt.Errorf("%w: error parsing json", err)
// 	}
// 	return nil
// }
