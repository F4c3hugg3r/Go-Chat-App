package chat

import (
	"fmt"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

type WebRTCInterface interface {
	Execute(message *ty.Message) (*ty.Response, error)
}

type WebRTCRegistry struct {
	plugins map[string]WebRTCInterface
}

// RegisterPlugins sets up all the plugins
func RegisterCallPlugins(chatService *ChatService) *WebRTCRegistry {
	cr := &WebRTCRegistry{plugins: make(map[string]WebRTCInterface)}
	cr.plugins[fmt.Sprint("/", ty.InitializeSignalFlag)] = NewInitializeSignalPluginPlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.OfferSignalFlag)] = NewOfferSignalPlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.AnswerSignalFlag)] = NewAnswerSignalPlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.ICECandidateFlag)] = NewICECandidatePlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.StableSignalFlag)] = NewStableSignalPlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.ConnectedFlag)] = NewConnectedPlugin(chatService)
	cr.plugins[fmt.Sprint("/", ty.FailedConnectionFlag)] = NewFailedConnectionPlugin(chatService)

	return cr
}

func (pr *WebRTCRegistry) FindAndExecute(message *ty.Message) (*ty.Response, error) {
	plugin, ok := pr.plugins[message.Plugin]
	if !ok {
		return &ty.Response{Err: fmt.Sprintf("%v: no such call plugin found: %s", ty.ErrNoPermission, message.Plugin)}, nil
	}

	return plugin.Execute(message)
}
