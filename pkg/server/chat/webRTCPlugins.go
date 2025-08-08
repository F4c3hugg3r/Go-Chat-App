package chat

import (
	"fmt"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

//
// IMPORTANT NOTE: for WebRTC Signals, Message.Name represents the ownId and Message.ClientId represents the oppId
//

// InitializeSignalPlugin initializes the rtc connection in the group and at the clients
type InitializeSignalPlugin struct {
	chatService *ChatService
}

func NewInitializeSignalPluginPlugin(s *ChatService) *InitializeSignalPlugin {
	return &InitializeSignalPlugin{chatService: s}
}

func (isp *InitializeSignalPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	fmt.Printf("\n[InitializeSignalPlugin] Execute called with msg: %+v", msg)
	group, ownClient, err := GetCurrentGroup(msg.Name, isp.chatService)
	if err != nil {
		fmt.Printf("\n[InitializeSignalPlugin] Error getting current group: %v", err)
		return nil, err
	}

	if group == nil {
		fmt.Printf("\n[InitializeSignalPlugin] Group is nil for msg.Name: %s", msg.Name)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	oppClient, err := isp.chatService.GetClient(msg.ClientId)
	if err != nil {
		fmt.Printf("\n[InitializeSignalPlugin] Error getting opposing client: %v", err)
		return nil, fmt.Errorf("%w: error getting opposing client", err)
	}

	fmt.Printf("\n[InitializeSignalPlugin] checking if there is already a connection")
	if group.CheckConnection(msg.Name, msg.ClientId) {
		return nil, fmt.Errorf("%w: there is already a connection between %s - %s", ty.ErrNoPermission, msg.Name, msg.ClientId)
	}

	fmt.Printf("\n[InitializeSignalPlugin] no current connectin found, connecting %s - %s", msg.Name, msg.ClientId)
	group.SetConnection(msg.Name, msg.ClientId, false)

	err = ownClient.SetCallState(msg.ClientId, ty.OfferSignalFlag)
	if err != nil {
		return nil, err
	}

	err = oppClient.SetCallState(msg.Name, ty.AnswerSignalFlag)

	return nil, err
}

// OfferSignalPlugin forwards rtc signals
type OfferSignalPlugin struct {
	chatService *ChatService
}

func NewOfferSignalPlugin(s *ChatService) *OfferSignalPlugin {
	return &OfferSignalPlugin{chatService: s}
}

func (osp *OfferSignalPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	fmt.Printf("\n[OfferSignalPlugin] Execute called with msg: %+v", msg)
	group, ownClient, err := GetCurrentGroup(msg.Name, osp.chatService)
	if err != nil {
		fmt.Printf("\n[OfferSignalPlugin] Error getting current group: %v", err)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if group == nil {
		fmt.Printf("\n[OfferSignalPlugin] Group is nil for msg.Name: %s", msg.Name)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	oppClient, err := osp.chatService.GetClient(msg.ClientId)
	if err != nil {
		fmt.Printf("\n[OfferSignalPlugin] Error getting opposing client: %v", err)
		return nil, fmt.Errorf("%w: error getting opposing client", err)
	}

	if msg.Content == "" {
		fmt.Printf("\n[OfferSignalPlugin] Empty content, setting call state to OfferSignalFlag for client %s", msg.ClientId)
		err = ownClient.SetCallState(msg.ClientId, ty.OfferSignalFlag)
		return nil, err
	}

	if ownClient.GetCallState(msg.ClientId) != ty.StableSignalFlag &&
		ownClient.GetCallState(msg.ClientId) != ty.OfferSignalFlag &&
		oppClient.GetCallState(msg.Name) != ty.StableSignalFlag &&
		oppClient.GetCallState(msg.Name) != ty.AnswerSignalFlag {

		fmt.Printf("\n[OfferSignalPlugin] Wrong call state: ownClient: %s, oppClient: %s", ownClient.GetCallState(msg.ClientId), oppClient.GetCallState(msg.Name))
		return nil, fmt.Errorf("%w: Offer couldn't be sent, because ownclient %s or oppClient %s"+
			"is in the wrong callState", ty.ErrNoPermission, ownClient.GetCallState(msg.ClientId), oppClient.GetCallState(msg.Name))
	}

	fmt.Printf("\n[OfferSignalPlugin] Forwarding Offer from %s to %s", msg.Name, msg.ClientId)
	osp.chatService.ForwardSignal(msg, ty.OfferSignalFlag)

	return nil, err
}

// AnswerSignalPlugin forwards rtc signals
type AnswerSignalPlugin struct {
	chatService *ChatService
}

func NewAnswerSignalPlugin(s *ChatService) *AnswerSignalPlugin {
	return &AnswerSignalPlugin{chatService: s}
}

func (asp *AnswerSignalPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	fmt.Printf("\n[AnswerSignalPlugin] Execute called with msg: %+v", msg)
	group, ownClient, err := GetCurrentGroup(msg.Name, asp.chatService)
	if err != nil {
		fmt.Printf("\n[AnswerSignalPlugin] Error getting current group: %v", err)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}
	if group == nil {
		fmt.Printf("\n[AnswerSignalPlugin] Group is nil for msg.Name: %s", msg.Name)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	err = ownClient.SetCallState(msg.ClientId, ty.AnswerSignalFlag)
	if err != nil {
		return nil, err
	}

	if msg.Content == "" {
		fmt.Printf("\n[AnswerSignalPlugin] Empty content, setting call state to AnswerSignalFlag for client %s", msg.ClientId)
		return nil, nil
	}

	oppClient, err := asp.chatService.GetClient(msg.ClientId)
	if err != nil {
		fmt.Printf("\n[AnswerSignalPlugin] Error getting opposing client: %v", err)
		return nil, fmt.Errorf("%w: error getting opposing client", err)
	}

	if ownClient.GetCallState(msg.ClientId) != ty.AnswerSignalFlag &&
		oppClient.GetCallState(msg.ClientId) != ty.OfferSignalFlag {

		fmt.Printf("\n[AnswerSignalPlugin] Wrong call state: ownClient: %s, oppClient: %s", ownClient.GetCallState(msg.ClientId), oppClient.GetCallState(msg.ClientId))
		return nil, fmt.Errorf("%w: offer couldn't be sent, because oppClient"+
			"is in the wrong callState %s", ty.ErrNoPermission, ownClient.GetCallState(msg.ClientId))
	}

	// maybe TODO prolly überflüssig wie brücken
	// no current connection so connection is canceled
	// if !group.CheckConnection(msg.Name, msg.ClientId) {
	// 	fmt.Printf("\n[AnswerSignalPlugin] No registered connection between %s and %s", msg.Name, msg.ClientId)
	// 	return nil, fmt.Errorf("%w: offer couldn't be sent because there is no registered connection", ty.ErrNotAvailable)
	// }

	fmt.Printf("\n[AnswerSignalPlugin] Forwarding answer signal between %s and %s", msg.Name, msg.ClientId)
	asp.chatService.ForwardSignal(msg, ty.AnswerSignalFlag)

	return nil, nil
}

// ICECandidatePlugin forwards rtc signals
type ICECandidatePlugin struct {
	chatService *ChatService
}

func NewICECandidatePlugin(s *ChatService) *ICECandidatePlugin {
	return &ICECandidatePlugin{chatService: s}
}

func (ice *ICECandidatePlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	fmt.Printf("\n[ICECandidatePlugin] Execute called with msg: %+v", msg)
	group, _, err := GetCurrentGroup(msg.Name, ice.chatService)
	if err != nil {
		fmt.Printf("\n[ICECandidatePlugin] Error getting current group: %v", err)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if group == nil {
		fmt.Printf("\n[ICECandidatePlugin] Group is nil for msg.Name: %s", msg.Name)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if !group.CheckConnection(msg.Name, msg.ClientId) {
		fmt.Printf("\n[ICECandidatePlugin] No registered connection between %s and %s", msg.Name, msg.ClientId)
		return nil, fmt.Errorf("%w: offer couldn't be sent because there is no registered connection", ty.ErrNotAvailable)
	}

	fmt.Printf("\n[ICECandidatePlugin] Forwarding ICE candidate between %s and %s", msg.Name, msg.ClientId)
	ice.chatService.ForwardSignal(msg, ty.ICECandidateFlag)

	return nil, nil
}

// StableSignalPlugin forwards rtc signals
type StableSignalPlugin struct {
	chatService *ChatService
}

func NewStableSignalPlugin(s *ChatService) *StableSignalPlugin {
	return &StableSignalPlugin{chatService: s}
}

func (ssp *StableSignalPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	fmt.Printf("\n[StableSignalPlugin] Execute called with msg: %+v", msg)
	ownClient, err := ssp.chatService.GetClient(msg.Name)
	if err != nil {
		fmt.Printf("\n[StableSignalPlugin] Error getting current group: %v", err)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	fmt.Printf("\n[StableSignalPlugin] Setting call state to StableSignalFlag for client %s", msg.ClientId)
	err = ownClient.SetCallState(msg.ClientId, ty.StableSignalFlag)
	return nil, err
}

// ConnectedPlugin forwards rtc signals
type ConnectedPlugin struct {
	chatService *ChatService
}

func NewConnectedPlugin(s *ChatService) *ConnectedPlugin {
	return &ConnectedPlugin{chatService: s}
}

func (cp *ConnectedPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	fmt.Printf("\n[ConnectedPlugin] Execute called with msg: %+v", msg)
	group, ownClient, err := GetCurrentGroup(msg.Name, cp.chatService)
	if err != nil {
		fmt.Printf("\n[ConnectedPlugin] Error getting current group: %v", err)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if group == nil {
		fmt.Printf("\n[ConnectedPlugin] Group is nil for msg.Name: %s", msg.Name)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	fmt.Printf("\n[ConnectedPlugin] Setting call state to ConnectedFlag for client %s and establishing connection", msg.ClientId)
	err = ownClient.SetCallState(msg.ClientId, ty.ConnectedFlag)
	group.SetConnection(msg.Name, msg.ClientId, true)

	return nil, err
}

// FailedConnectionPlugin forwards rtc signals
type FailedConnectionPlugin struct {
	chatService *ChatService
}

func NewFailedConnectionPlugin(s *ChatService) *FailedConnectionPlugin {
	return &FailedConnectionPlugin{chatService: s}
}

func (fcp *FailedConnectionPlugin) Execute(msg *ty.Message) (*ty.Response, error) {
	fmt.Printf("\n[FailedConnectionPlugin] Execute called with msg: %+v", msg)
	group, ownClient, err := GetCurrentGroup(msg.Name, fcp.chatService)
	if err != nil {
		fmt.Printf("\n[FailedConnectionPlugin] Error getting current group: %v", err)
		return nil, fmt.Errorf("%w: error getting current group", err)
	}

	if msg.ClientId == "" {
		fmt.Printf("\n[FailedConnectionPlugin] ClientId is empty, removing unconnected RTCs for %s", msg.Name)
		ownClient.RemoveUnconnectedRTCs()
		return nil, nil
	}

	oppClient, err := fcp.chatService.GetClient(msg.ClientId)
	if err != nil {
		fmt.Printf("\n[FailedConnectionPlugin] Error getting opposing client: %v", err)
		return nil, fmt.Errorf("%w: error getting opposing client", err)
	}

	if msg.Content == ty.RollbackDoneFlag {
		fmt.Printf("\n[FailedConnectionPlugin] Rollback done for %s and %s", msg.Name, msg.ClientId)
		if group != nil {
			group.RemoveConnection(msg.Name, msg.ClientId, false)
		}
		ownClient.RemoveRTC(msg.ClientId)
		oppClient.RemoveRTC(msg.Name)
		return nil, nil
	}

	fmt.Printf("\n[FailedConnectionPlugin] Echoing failed connection to %s and %s", msg.Name, msg.ClientId)
	fcp.chatService.Echo(msg.Name, &ty.Response{ClientId: msg.ClientId, RspName: ty.FailedConnectionFlag, Content: ty.FailedConnectionFlag})
	fcp.chatService.Echo(msg.ClientId, &ty.Response{ClientId: msg.Name, RspName: ty.FailedConnectionFlag, Content: ty.FailedConnectionFlag})

	return nil, nil
}
