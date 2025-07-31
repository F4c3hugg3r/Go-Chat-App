package chat

import (
	"fmt"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// OfferSignalPlugin forwards rtc signals
type OfferSignalPlugin struct {
	chatService *ChatService
}

func NewOfferSignalPlugin(s *ChatService) *OfferSignalPlugin {
	return &OfferSignalPlugin{chatService: s}
}

// TODO Anruf annehmen / ablehnen
func (cp *OfferSignalPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	group, ownClient, err := GetCurrentGroup(msg.Name, cp.chatService)
	if err != nil {
		return nil, fmt.Errorf("%w: error getting current group", err)
	}
	if group == nil {
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if msg.Content == "" {
		ownClient.SetCallState(msg.ClientId, ty.OfferSignalFlag)
		return nil, nil
	}

	oppClient, err := cp.chatService.GetClient(msg.ClientId)
	if err != nil {
		return nil, fmt.Errorf("%w: error getting opposing client", err)
	}

	if ownClient.GetCallState(msg.ClientId) != ty.StableSignalFlag &&
		ownClient.GetCallState(msg.ClientId) != ty.OfferSignalFlag &&
		oppClient.GetCallState(msg.Name) != ty.StableSignalFlag &&
		oppClient.GetCallState(msg.Name) != ty.AnswerSignalFlag {

		return nil, fmt.Errorf("%w: Offer couldn't be sent, because ownclient %s or oppClient %s"+
			"is in the wrong callState", ty.ErrNoPermission, ownClient.GetCallState(msg.ClientId), oppClient.GetCallState(msg.Name))
	}

	// die manuelle callState Zuweisung (OfferSignal) von ownClient
	// ist nicht nötig, da das schon in webrtc.go initiiert wird

	// check if connection is already established in group
	if group.CheckConnection(msg.Name, msg.ClientId) {
		cp.chatService.ForwardSignal(msg, ty.OfferSignalFlag)
		oppClient.SetCallState(msg.Name, ty.AnswerSignalFlag)

		return nil, nil
	}

	// no current connection so connection is established in group
	group.SetConnection(msg.Name, msg.ClientId, false)
	cp.chatService.ForwardSignal(msg, ty.OfferSignalFlag)
	oppClient.SetCallState(msg.Name, ty.AnswerSignalFlag)

	return nil, nil
}

// AnswerSignalPlugin forwards rtc signals
type AnswerSignalPlugin struct {
	chatService *ChatService
}

func NewAnswerSignalPlugin(s *ChatService) *AnswerSignalPlugin {
	return &AnswerSignalPlugin{chatService: s}
}

func (cp *AnswerSignalPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	group, ownClient, err := GetCurrentGroup(msg.Name, cp.chatService)
	if err != nil {
		return nil, fmt.Errorf("%w: error getting current group", err)
	}
	if group == nil {
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if msg.Content == "" {
		ownClient.SetCallState(msg.ClientId, ty.AnswerSignalFlag)
		return nil, nil
	}

	oppClient, err := cp.chatService.GetClient(msg.ClientId)
	if err != nil {
		return nil, fmt.Errorf("%w: error getting opposing client", err)
	}

	if ownClient.GetCallState(msg.ClientId) != ty.AnswerSignalFlag &&
		oppClient.GetCallState(msg.ClientId) != ty.OfferSignalFlag {

		return nil, fmt.Errorf("%w: offer couldn't be sent, because oppClient"+
			"is in the wrong callState %s", ty.ErrNoPermission, ownClient.GetCallState(msg.ClientId))
	}

	// no current connection so connection is canceled
	if !group.CheckConnection(msg.Name, msg.ClientId) {
		return nil, fmt.Errorf("%w: offer couldn't be sent because there is no registered connection", ty.ErrNotAvailable)
	}

	cp.chatService.ForwardSignal(msg, ty.AnswerSignalFlag)

	return nil, nil
}

// ICECandidatePlugin forwards rtc signals
type ICECandidatePlugin struct {
	chatService *ChatService
}

func NewICECandidatePlugin(s *ChatService) *ICECandidatePlugin {
	return &ICECandidatePlugin{chatService: s}
}

func (cp *ICECandidatePlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	group, _, err := GetCurrentGroup(msg.Name, cp.chatService)
	if err != nil {
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if group == nil {
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if !group.CheckConnection(msg.Name, msg.ClientId) {
		return nil, fmt.Errorf("%w: offer couldn't be sent because there is no registered connection", ty.ErrNotAvailable)
	}

	cp.chatService.ForwardSignal(msg, ty.ICECandidateFlag)

	return nil, nil
}

// StableSignalPlugin forwards rtc signals
type StableSignalPlugin struct {
	chatService *ChatService
}

func NewStableSignalPlugin(s *ChatService) *StableSignalPlugin {
	return &StableSignalPlugin{chatService: s}
}

func (cp *StableSignalPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	ownClient, err := cp.chatService.GetClient(msg.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	ownClient.SetCallState(msg.ClientId, ty.StableSignalFlag)
	return nil, nil
}

// ConnectedPlugin forwards rtc signals
type ConnectedPlugin struct {
	chatService *ChatService
}

func NewConnectedPlugin(s *ChatService) *ConnectedPlugin {
	return &ConnectedPlugin{chatService: s}
}

func (cp *ConnectedPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	group, ownClient, err := GetCurrentGroup(msg.Name, cp.chatService)
	if err != nil {
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if group == nil {
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	ownClient.SetCallState(msg.ClientId, ty.ConnectedFlag)
	group.SetConnection(msg.Name, msg.ClientId, true)

	return nil, nil
}

// FailedConnectionPlugin forwards rtc signals
type FailedConnectionPlugin struct {
	chatService *ChatService
}

func NewFailedConnectionPlugin(s *ChatService) *FailedConnectionPlugin {
	return &FailedConnectionPlugin{chatService: s}
}

func (cp *FailedConnectionPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	group, ownClient, err := GetCurrentGroup(msg.Name, cp.chatService)
	if err != nil {
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if msg.ClientId == "" {
		ownClient.RemoveUnconnectedRTCs()
		return nil, nil
	}

	oppClient, err := cp.chatService.GetClient(msg.ClientId)
	if err != nil {
		return nil, fmt.Errorf("%w: error getting opposing client", err)
	}

	if msg.Content == ty.RollbackDoneFlag {
		if group != nil {
			group.RemoveConnection(msg.Name, msg.ClientId, false)
		}
		ownClient.RemoveRTC(msg.ClientId)
		oppClient.RemoveRTC(msg.Name)
		return nil, nil
	}

	cp.chatService.Echo(msg.Name, &ty.Response{ClientId: msg.ClientId, RspName: ty.FailedConnectionFlag, Content: ty.FailedConnectionFlag})
	cp.chatService.Echo(msg.ClientId, &ty.Response{ClientId: msg.Name, RspName: ty.FailedConnectionFlag, Content: ty.FailedConnectionFlag})

	return nil, nil
}
