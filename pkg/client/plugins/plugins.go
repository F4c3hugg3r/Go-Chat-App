package plugins

import (
	"encoding/json"
	"fmt"
	"strings"

	n "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/network"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// TODO Option für Mic/Speaker refresh

// CallPlugin lets you participate in a voice call
type CallPlugin struct {
	c *n.Client
}

func NewCallPlugin(chatClient *n.Client) *CallPlugin {
	return &CallPlugin{c: chatClient}
}

func (cp *CallPlugin) CheckScope() int {
	return RegisteredOnly
}

func (cp *CallPlugin) Execute(message *t.Message) (error, string) {
	cp.c.LogChan <- t.Log{Text: "CallPlugin.Execute: Initiating call plugin execution", Method: "CallPlugin.Execute"}

	switch {
	case strings.Contains(message.Content, "quit"):
		cp.c.LogChan <- t.Log{Text: "CallPlugin.Execute: Received 'quit' command, deleting peers", Method: "CallPlugin.Execute"}
		cp.c.DeletePeersSafely(cp.c.GetClientId(), true, true)
		return nil, ""

	case strings.Contains(message.Content, "accept"):
		cp.c.LogChan <- t.Log{Text: "CallPlugin.Execute: Received 'accept' command, answering call initialization as accepted", Method: "CallPlugin.Execute"}
		cp.c.CallTimeoutChan <- true
		return cp.c.AnswerCallInitialization(message, t.CallAccepted), ""

	case strings.Contains(message.Content, "deny"):
		cp.c.LogChan <- t.Log{Text: "CallPlugin.Execute: Received 'deny' command, answering call initialization as denied", Method: "CallPlugin.Execute"}
		cp.c.CallTimeoutChan <- false
		return cp.c.AnswerCallInitialization(message, t.CallDenied), ""
	}

	// gathering group clients
	cp.c.LogChan <- t.Log{Text: fmt.Sprintf("Posting message to server: %+v", message), Method: "CallPlugin.Execute"}
	rsp, err := cp.c.PostMessage(message, t.PostPlugin)
	if err != nil || rsp.Err != t.IgnoreResponseTag {
		cp.c.LogChan <- t.Log{Text: fmt.Sprintf("Error posting or processing response: %v", err), Method: "CallPlugin.Execute"}
		return err, ""
	}

	cp.c.LogChan <- t.Log{Text: fmt.Sprintf("Received response: %s", rsp.Content), Method: "CallPlugin.Execute"}

	var callableClientIds []string
	dec := json.NewDecoder(strings.NewReader(rsp.Content))
	err = dec.Decode(&callableClientIds)
	if err != nil {
		cp.c.LogChan <- t.Log{Text: fmt.Sprintf("Error decoding client IDs from response: %s", rsp.Content), Method: "CallPlugin.Execute"}
		cp.c.SendSignalingError("", message.ClientId, "")
		return err, ""
	}
	cp.c.LogChan <- t.Log{Text: fmt.Sprintf("Decoded client IDs: %s", strings.Join(callableClientIds, ", ")), Method: "CallPlugin.Execute"}

	for _, oppClientId := range callableClientIds {
		cp.c.LogChan <- t.Log{Text: fmt.Sprintf("Starting HandleSignal for client %s", oppClientId), Method: "CallPlugin.Execute"}
		go cp.c.HandleSignal(&t.Response{ClientId: oppClientId}, true, false)
	}

	cp.c.LogChan <- t.Log{Text: "CallPlugin.Execute: Finished sending initialisation offers", Method: "CallPlugin.Execute"}
	return nil, ""
}

// GroupPlugin lets you participate in a group chat
type GroupPlugin struct {
	c *n.Client
}

func NewGroupPlugin(chatClient *n.Client) *GroupPlugin {
	return &GroupPlugin{c: chatClient}
}

func (gp *GroupPlugin) CheckScope() int {
	return RegisteredOnly
}

func (gp *GroupPlugin) Execute(message *t.Message) (error, string) {
	// wird im responeEvaluator gemacht
	// if strings.Contains(message.Content, t.LeaveGroupFlag) {
	// 	gp.c.DeletePeersSafely(message.ClientId, true, true)
	// }
	_, err := gp.c.PostMessage(message, t.PostPlugin)
	return err, ""
}

// PrivateMessage Plugin lets a client send a private message to another client identified by it's clientId
type PrivateMessagePlugin struct {
	c *n.Client
}

func NewPrivateMessagePlugin(chatClient *n.Client) *PrivateMessagePlugin {
	return &PrivateMessagePlugin{c: chatClient}
}

func (pp *PrivateMessagePlugin) CheckScope() int {
	return RegisteredOnly
}

func (pp *PrivateMessagePlugin) Execute(message *t.Message) (error, string) {
	if message.Content == "" {
		return fmt.Errorf("%w: prefix shouldn't be empty", t.ErrParsing), ""
	}
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
	c *n.Client
}

func NewLogOutPlugin(chatClient *n.Client) *LogOutPlugin {
	return &LogOutPlugin{c: chatClient}
}

func (lp *LogOutPlugin) CheckScope() int {
	return RegisteredOnly
}

func (lp *LogOutPlugin) Execute(message *t.Message) (error, string) {
	// wird im responeEvaluator gemacht
	//lp.c.DeletePeersSafely(message.ClientId, true, true)
	return lp.c.PostDelete(message), t.UnregisterFlag
}

// RegisterClientPlugin safely registeres a client by creating a Client with the received values
// and putting it into the global clients map
type RegisterClientPlugin struct {
	c *n.Client
}

func NewRegisterClientPlugin(chatClient *n.Client) *RegisterClientPlugin {
	return &RegisterClientPlugin{c: chatClient}
}

func (rp *RegisterClientPlugin) CheckScope() int {
	return UnregisteredOnly
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

	return err, t.RegisterFlag
}

// BroadcaastPlugin distributes an incomming message abroad all client channels if
// a client can't receive, i'ts active status is set to false
type BroadcastPlugin struct {
	c *n.Client
}

func NewBroadcastPlugin(chatClient *n.Client) *BroadcastPlugin {
	return &BroadcastPlugin{c: chatClient}
}

func (bp *BroadcastPlugin) CheckScope() int {
	return RegisteredOnly
}

func (bp *BroadcastPlugin) Execute(message *t.Message) (error, string) {
	_, err := bp.c.PostMessage(message, t.PostPlugin)
	return err, ""
}

// HelpPlugin tells you information about available plugins
type HelpPlugin struct {
	c *n.Client
}

func NewHelpPlugin(chatClient *n.Client) *HelpPlugin {
	return &HelpPlugin{c: chatClient}
}

func (hp *HelpPlugin) CheckScope() int {
	return RegisteredOnly
}

func (h *HelpPlugin) Execute(message *t.Message) (error, string) {
	_, err := h.c.PostMessage(message, t.PostPlugin)
	return err, ""
}

// UserPlugin tells you information about all the current users
type UserPlugin struct {
	c *n.Client
}

func NewUserPlugin(chatClient *n.Client) *UserPlugin {
	return &UserPlugin{c: chatClient}
}

func (up *UserPlugin) CheckScope() int {
	return RegisteredOnly
}

func (u *UserPlugin) Execute(message *t.Message) (error, string) {
	_, err := u.c.PostMessage(message, t.PostPlugin)
	return err, ""
}

// TimePlugin tells you the current time
type TimePlugin struct {
	c *n.Client
}

func NewTimePlugin(chatClient *n.Client) *TimePlugin {
	return &TimePlugin{c: chatClient}
}

func (tp *TimePlugin) CheckScope() int {
	return RegisteredOnly
}

func (tp *TimePlugin) Execute(message *t.Message) (error, string) {
	_, err := tp.c.PostMessage(message, t.PostPlugin)
	return err, ""
}
