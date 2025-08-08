package UI

import (
	"fmt"
	"strings"

	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
	"github.com/charmbracelet/lipgloss"
)

// TODO das in Interface bzw Plugins auslagern zB rsp in Plugin verarbeiten
// und in error flag zurückgeben, dass es nicht geprinted werden soll

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

	// empty output
	case rsp.Content == "":
		return ""

	// register output
	case strings.Contains(rsp.Content, t.RegisterFlag):
		m.RenderTitle(t.RegisterFlag, []string{m.userService.ChatClient.GetName()})
		m.userService.Executor("/users")

		return purple.BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).
			BorderForeground(purple.GetForeground()).Render(blue.Render(RegisterOutput))

	// server output
	case rsp.RspName == "":
		rspString = fmt.Sprintf("%s", blue.Render(rsp.Content))

		// unregister output
		if strings.Contains(rsp.Content, t.UnregisterFlag) {
			m.RenderTitle(t.UnregisterFlag, nil)
			m.userService.ChatClient.DeletePeersSafely("", true, true)
			m.userService.ChatClient.ClientChangeSignalChan <- t.ClientsChangeSignal{
				CallState: t.UnregisterFlag,
			}
		}

		return rspString

	// one user left output
	case strings.Contains(rsp.RspName, t.UserRemoveFlag):
		return fmt.Sprintf("%s %s", purple.Render(rsp.Content), blue.Faint(true).Render("hat den Chat verlassen"))

	// one user joined output
	case strings.Contains(rsp.RspName, t.UserAddFlag):
		return fmt.Sprintf("%s %s", purple.Render(rsp.Content), blue.Faint(true).Render("ist dem Chat beigetreten"))

	// addGroup output
	case strings.Contains(rsp.RspName, t.AddGroupFlag):
		group, err := m.userService.HandleAddGroup(rsp.Content)
		if err != nil {
			return red.Render(fmt.Sprintf("%v: error formatting json to group", err))
		}

		m.RenderTitle(t.AddGroupFlag, []string{m.userService.ChatClient.GetName(), group.Name})
		m.userService.Executor("/group users")

		return purple.BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).
			BorderForeground(purple.GetForeground()).Render(fmt.Sprintf("%s %s\n%s",
			blue.Render("-> Du bist nun Teil der Gruppe"),
			turkis.Render(group.Name),
			blue.Faint(true).Render("Private Nachrichten kannst du weiterhin außerhalb verschicken"),
		))

	// leaveGroup output
	case strings.Contains(rsp.RspName, t.LeaveGroupFlag):
		m.userService.ChatClient.UnsetGroupId()
		m.RenderTitle(t.RegisterFlag, []string{m.userService.ChatClient.GetName()})
		m.userService.ChatClient.DeletePeersSafely("", true, true)
		m.userService.Executor("/users")

		return purple.BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).
			BorderForeground(purple.GetForeground()).
			Render(blue.Render(RegisterOutput))

	// Rollback/Delete Peer output
	case strings.Contains(rsp.RspName, t.FailedConnectionFlag):
		m.userService.ChatClient.DeletePeersSafely(rsp.ClientId, false, false)
		return blue.Faint(true).Render(fmt.Sprintf("Anruf mit %s beendet", purple.Faint(true).Render(rsp.ClientId)))

	// Receive webRTC signal (Offer SDP Signal, Answer SDP Signal or ICE Candidate)
	case strings.Contains(rsp.RspName, t.OfferSignalFlag),
		strings.Contains(rsp.RspName, t.AnswerSignalFlag),
		strings.Contains(rsp.RspName, t.ICECandidateFlag):

		m.logChan <- t.Log{Text: "webrtc related response detected"}
		m.userService.ChatClient.HandleSignal(rsp, false)

		return ""

	// slice output
	case strings.HasPrefix(rsp.Content, "["):
		if strings.Contains(rsp.RspName, t.UsersFlag) {
			m.userService.ChatClient.ClientChangeSignalChan <- t.ClientsChangeSignal{
				ClientsJson: rsp.Content,
			}
			return ""
		}

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
