package UI

import (
	"fmt"
	"strings"

	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
	"github.com/charmbracelet/lipgloss"
)

// evaluateResponse evaluates an incoming Response and returns the
// corresponding rendered string
func (m *model) EvaluateReponse(rsp *t.Response) string {
	var rspString string

	switch {
	// error output
	case rsp.Err != "":
		if rsp.Err == t.IgnoreResponseTag {
			return ""
		}
		return red.Render(rsp.Err)

	// Users output
	case strings.Contains(rsp.RspName, t.UsersFlag):
		m.userService.Client.ClientChangeSignalChan <- t.ClientsChangeSignal{
			ClientsJson: rsp.Content,
		}
		m.logChan <- t.Log{Text: fmt.Sprintf("clients = %s", rsp.Content)}

		return ""

	// empty output
	case rsp.Content == "", rsp.Content == "null":
		return ""

	// register output
	case strings.Contains(rsp.Content, t.RegisterFlag):
		m.RenderTitle(t.RegisterFlag, []string{m.userService.Client.GetName()})
		m.userService.Executor("/users")

		return purple.BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).
			BorderForeground(purple.GetForeground()).Render(blue.Render(RegisterOutput))

	// server output
	case rsp.RspName == "":
		rspString = fmt.Sprintf("%s", blue.Render(rsp.Content))

		// unregister output
		if strings.Contains(rsp.Content, t.UnregisterFlag) {
			m.RenderTitle(t.UnregisterFlag, nil)
			m.userService.Client.DeletePeers("", true, true)
			m.userService.Client.ClientChangeSignalChan <- t.ClientsChangeSignal{
				CallState: t.UnregisterFlag,
			}
		}

		return rspString

	// one user left output
	case strings.Contains(rsp.RspName, t.UserRemoveFlag):
		if m.userService.Client.GetGroupId() != "" {
			m.userService.Executor("/group users")
		} else {
			m.userService.Executor("/users")
		}
		return fmt.Sprintf("%s %s", purple.Render(rsp.Content), blue.Faint(true).Render("hat den Chat verlassen"))

	// one user joined output
	case strings.Contains(rsp.RspName, t.UserAddFlag):
		if m.userService.Client.GetGroupId() != "" {
			m.userService.Executor("/group users")
		} else {
			m.userService.Executor("/users")
		}
		return fmt.Sprintf("%s %s", purple.Render(rsp.Content), blue.Faint(true).Render("ist dem Chat beigetreten"))

	// addGroup output
	case strings.Contains(rsp.RspName, t.AddGroupFlag):
		group, err := m.userService.HandleAddGroup(rsp.Content)
		if err != nil {
			return red.Render(fmt.Sprintf("%v: error formatting json to group", err))
		}

		m.RenderTitle(t.AddGroupFlag, []string{m.userService.Client.GetName(), group.Name})
		m.userService.Executor("/group users")

		return purple.BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).
			BorderForeground(purple.GetForeground()).Render(fmt.Sprintf("%s %s\n%s",
			blue.Render("-> Du bist nun Teil der Gruppe"),
			turkis.Render(group.Name),
			blue.Faint(true).Render("Private Nachrichten kannst du weiterhin außerhalb verschicken"),
		))

	// leaveGroup output
	case strings.Contains(rsp.RspName, t.LeaveGroupFlag):
		m.userService.Client.UnsetGroupId()
		m.RenderTitle(t.RegisterFlag, []string{m.userService.Client.GetName()})
		m.userService.Client.DeletePeers("", true, true)
		m.userService.Executor("/users")

		return purple.BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).
			BorderForeground(purple.GetForeground()).
			Render(blue.Render(RegisterOutput))

	// Rollback/Delete Peer output
	case strings.Contains(rsp.RspName, t.FailedConnectionFlag):
		if len(m.userService.Client.Peers) < 1 {
			return ""
		}

		m.userService.Client.DeletePeers(rsp.ClientId, false, false)
		if m.userService.Client.GetGroupId() != "" {
			m.userService.Executor("/group users")
		} else {
			m.userService.Executor("/users")
		}

		return fmt.Sprintf("%s %s %s", blue.Render("Anruf mit"), purple.Faint(true).Render(rsp.ClientId), blue.Render("beendet"))

	// Receive webRTC signal (Offer SDP Signal, Answer SDP Signal or ICE Candidate)
	case strings.Contains(rsp.RspName, t.OfferSignalFlag),
		strings.Contains(rsp.RspName, t.AnswerSignalFlag),
		strings.Contains(rsp.RspName, t.ICECandidateFlag),
		strings.Contains(rsp.RspName, t.InitializeSignalFlag):
		m.logChan <- t.Log{Text: "webrtc related signal detected"}

		switch rsp.Content {
		case t.ReceiveCall:
			return fmt.Sprintf("%s%s", green.Render("Du wirst angerufen! Antworte innerhalb von 15sek:"),
				blue.Render("\n		-> '/call accept' um anzunehmen"+
					"\n		-> '/call deny' um abzulehnen"))

		case t.CallAccepted:
			m.userService.Client.HandleSignal(rsp, false, true)
			return green.Render("Dein Anruf wurde angenommen, verbinde...")

		case t.CallDenied:
			m.userService.Client.DeletePeers(rsp.ClientId, false, false)
			return green.Render("- Dein Anruf wurde abgelehnt -")
		}

		m.userService.Client.HandleSignal(rsp, false, false)

		return ""

	// slice output
	case strings.HasPrefix(rsp.Content, "["):

		output, err := JSONToTable(rsp.Content)
		if err != nil {
			return red.Render(fmt.Sprintf("%v: error formatting json to table", err))
		}

		return output
	}

	// response output
	rspString = fmt.Sprintf("%s: %s", turkis.Render(rsp.RspName), rsp.Content)

	return rspString
}
