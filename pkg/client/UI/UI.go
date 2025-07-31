package UI

import (
	"fmt"
	"strings"

	i "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/input"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TODO ALLGEMEIN

// TODO Darstellung des Anrufes in UI

// InitialModel initializes the model struct, which is the main struct for the TUI
func InitialModel(u *i.UserService) model {
	ti := SetUpTextInput(u)

	h := help.New()

	ta, tV := setUpTable()

	vp := viewport.New(30, 5)
	vp.KeyMap = viewportKeys

	logVp := viewport.New(30, 0)

	inputManager := &InputHistory{
		current: -1,
		inputs:  make([]string, 0, 200),
	}

	model := model{
		messages:    []string{},
		viewport:    vp,
		logViewport: logVp,
		logs:        []string{},
		err:         nil,
		userService: u,
		outputChan:  u.ChatClient.Output,
		textinput:   ti,
		help:        h,
		keyMap:      helpKeys,
		inH:         inputManager,
		title:       UnregisterTitle,
		table:       ta,
		tableValues: tV,
	}

	model.logChan = model.userService.ChatClient.LogChan

	return model
}

// Init is being called before Update listenes and initializes required functions
func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.waitForExternalResponse(), m.waitForLog())
}

// Update handles every input
func (m model) Update(rsp tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		tiCmd tea.Cmd
		loCmd tea.Cmd
	)

	m.viewport, vpCmd = m.viewport.Update(rsp)
	m.textinput, tiCmd = m.textinput.Update(rsp)
	m.logViewport, loCmd = m.logViewport.Update(rsp)

	if m.textinput.Value() == "" {
		m.inH.SaveInput("")
	}

	switch rsp := rsp.(type) {
	case t.Log:
		m.PrintLog(rsp)

		return m, tea.Batch(tiCmd, vpCmd, loCmd, m.waitForLog())

	case *t.Response:
		m.HandleResponse(rsp)

		return m, tea.Batch(tiCmd, vpCmd, loCmd, m.waitForExternalResponse())

	case tea.WindowSizeMsg:
		m.HandleWindowResize(&rsp)
		m.logChan <- t.Log{Text: "log started"}

	case tea.KeyMsg:
		switch rsp.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.userService.ChatClient.Interrupt()

			return m, tea.Quit

		case tea.KeyEnter:
			m.HandleMessage()
		}

		switch {
		case key.Matches(rsp, m.keyMap.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(rsp, m.keyMap.Logs):
			m.HideLogs()

		case key.Matches(rsp, m.keyMap.InputLeft), key.Matches(rsp, m.keyMap.InputRight):
			input := m.SearchInputHistory(rsp)
			m.textinput.SetValue(input)
			m.textinput.CursorEnd()
		}

	case errMsg:
		m.err = rsp
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd, loCmd)
}

// View describes the terminal view
func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s%s%s%s%s%s%s",
		m.title,
		Gap,
		lipgloss.JoinHorizontal(lipgloss.Center, m.viewport.View(), m.table.View()),
		Gap,
		m.textinput.View(),
		Gap,
		m.help.View(m.keyMap),
		Gap,
		m.logViewport.View(),
	)
}

func (m *model) refreshLogViewPort() {
	if len(m.logs) > 0 {
		str, _ := strings.CutSuffix(strings.Join(m.logs, "\n"), "\n")
		m.logViewport.SetContent(lipgloss.NewStyle().Width(m.logViewport.Width).Render(str))
	}

	m.logViewport.GotoBottom()
}

func (m *model) refreshTable() {
	m.tableValues.ConvertClientsToRows()
	m.table.SetRows(m.tableValues.rows)

	// TODO wenn table nicht focussed ist, das styling ändern, sodass
	// die zeile nicht hervorsticht

	// TODO call events anzeigen lassen und evtl automatische Aktualisierung

	// TODO focussing anpassen
	// m.table.SetCursor(-1)
}

// refreshViewPort refreshes the size of the viewport
func (m *model) refreshViewPort() {
	if len(m.messages) > 0 {
		str, _ := strings.CutSuffix(strings.Join(m.messages, "\n"), "\n")
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(str))
	}

	m.viewport.GotoBottom()
}

func (m *model) HideLogs() {
	switch m.logViewport.Height {
	case 0:
		m.logViewport.Height = 6
		m.refreshLogViewPort()
	case 6:
		m.logViewport.Height = 0
	}
}

// SearchInputHistory takes a keybind for left or right, if needed, sets the current index value
// and decides which index of the history inputs to present
func (m *model) SearchInputHistory(rsp tea.KeyMsg) string {
	var pending int
	if len(m.inH.inputs) < 1 {
		return ""
	}

	first := m.inH.checkFirst()

	switch {
	case key.Matches(rsp, m.keyMap.InputRight):
		pending = m.inH.current + 1

	case key.Matches(rsp, m.keyMap.InputLeft):
		if !first {
			pending = m.inH.current - 1
		} else {
			pending = m.inH.current
		}
	}

	m.inH.setCurrentHistoryIndex(pending)

	return m.inH.inputs[m.inH.current]
}

// renderTitle decides between the registered and unregistered title sets it into the viewport
// and return the heightDiff of the old and new title
func (m *model) renderTitle(title string, param []string) {
	if param == nil || param[0] != WindowResizeFlag {
		switch {
		case strings.Contains(title, t.UnregisterFlag):
			title = UnregisterTitle

		case strings.Contains(title, t.RegisterFlag):
			title = titleStyle.Render(fmt.Sprintf(RegisterTitle,
				turkis.Render(param[0])))

		case strings.Contains(title, t.AddGroupFlag):
			title = titleStyle.Render(fmt.Sprintf(GroupTitle, turkis.Render(param[0]), turkis.Render(param[1])))
		}
	}

	heightDiff := lipgloss.Height(title) - lipgloss.Height(m.title)
	m.title = centered.Width(m.viewport.Width + m.table.Width()).Bold(true).Render(title)
	m.viewport.Height = m.viewport.Height - heightDiff
	m.table.SetHeight(m.table.Height() - heightDiff)
}

// HandleWindowResize handles rezising of the terminal window by updating all models sizes
func (m *model) HandleWindowResize(rsp *tea.WindowSizeMsg) {
	m.viewport.Width = rsp.Width / 5 * 4
	m.table.SetWidth(rsp.Width / 5)
	m.textinput.Width = rsp.Width
	m.help.Width = rsp.Width
	m.logViewport.Width = rsp.Width
	var logPortHeight int
	switch m.logViewport.Height {
	case 0:
		logPortHeight = 1
	default:
		logPortHeight = m.logViewport.Height
	}

	m.viewport.Height = rsp.Height - (lipgloss.Height(Gap) * 2) - lipgloss.Height(m.title) - logPortHeight
	m.table.SetHeight(rsp.Height - (lipgloss.Height(Gap) * 2) - lipgloss.Height(m.title) - logPortHeight)

	m.renderTitle(m.title, []string{WindowResizeFlag})
	m.refreshViewPort()
}

// HandleResponse handles an incoming Response by evaluating it and refreshing
// the viewport by adding the corresponding string
func (m *model) HandleResponse(rsp *t.Response) {
	str := m.evaluateReponse(rsp)
	if str != "" {
		m.messages = append(m.messages, str)
	}

	m.refreshViewPort()
}

// HandleMessage hanbles the input message by saving it for the inputHistory and
// executeing the fitting userService method
func (m *model) HandleMessage() {
	m.inH.SaveInput(m.textinput.Value())

	m.userService.Executor(m.textinput.Value())
	m.textinput.Reset()

	m.refreshViewPort()
}

// TODO das in Interface bzw Plugins auslagern zB rsp in Plugin verarbeiten
// und in error flag zurückgeben, dass es nicht geprinted werden soll

// evaluateResponse evaluates an incoming Response and returns the
// corresponding rendered string
func (m *model) evaluateReponse(rsp *t.Response) string {
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
		m.renderTitle(t.RegisterFlag, []string{m.userService.ChatClient.GetName()})

		return blue.Render("-> Du kannst nun Nachrichten schreiben oder Commands ausführen\n'/help' → Befehle anzeigen\n'/quit' → Chat verlassen")

	// server output
	case rsp.RspName == "":
		rspString = fmt.Sprintf("%s", blue.Render(rsp.Content))

		// unregister output
		if strings.Contains(rsp.Content, t.UnregisterFlag) {
			m.renderTitle(t.UnregisterFlag, nil)
		}
		m.userService.ChatClient.DeletePeersSafely("", true)

		return rspString

	// addGroup output
	case strings.Contains(rsp.RspName, t.AddGroupFlag):
		group, err := m.userService.HandleAddGroup(rsp.Content)
		if err != nil {
			return red.Render(fmt.Sprintf("%v: error formatting json to group", err))
		}

		m.renderTitle(t.AddGroupFlag, []string{m.userService.ChatClient.GetName(), group.Name})

		return blue.Render(fmt.Sprintf("-> Du bist nun Teil der Gruppe %s und kannst Nachrichten in ihr schreiben\nPrivate Nachrichten kannst du weiterhin außerhalb verschicken", group.Name))

	// leaveGroup output
	case strings.Contains(rsp.RspName, t.LeaveGroupFlag):
		m.userService.ChatClient.UnsetGroupId()
		m.renderTitle(t.RegisterFlag, []string{m.userService.ChatClient.GetName()})
		m.userService.ChatClient.DeletePeersSafely("", true)

		return blue.Render("Du hast die Gruppe verlassen!\n-> Du kannst nun Nachrichten schreiben oder Commands ausführen\n'/help' → Befehle anzeigen\n'/quit' → Chat verlassen")

	// Rollback/Delete Peer output
	case strings.Contains(rsp.RspName, t.FailedConnectionFlag):
		m.userService.ChatClient.DeletePeersSafely(rsp.ClientId, false)

	// Receive webRTC signal (Offer SDP Signal, Answer SDP Signal or ICE Candidate)
	case strings.Contains(rsp.RspName, t.OfferSignalFlag),
		strings.Contains(rsp.RspName, t.AnswerSignalFlag),
		strings.Contains(rsp.RspName, t.ICECandidateFlag):

		m.logChan <- t.Log{Text: "webrtc related response detected"}
		m.userService.ChatClient.HandleSignal(rsp, false)

		return ""

	// slice output
	case strings.HasPrefix(rsp.Content, "["):
		output, err := JSONToTable(rsp.Content)
		if err != nil {
			return red.Render(fmt.Sprintf("%v: error formatting json to table", err))
		}

		if strings.Contains(rsp.RspName, t.UsersFlag) {
			clients, err := ParseJsonToUIClients(rsp.Content)
			if err != nil {
				return red.Render(fmt.Sprintf("%v: error formatting json to clients", err))
			}

			// m.logChan <- t.Log{Text: fmt.Sprint("length clientslice after parsing: ", len(clients))}

			m.tableValues.SetClients(clients, nil)
			m.refreshTable()
		}

		return output
	}

	// response output
	rspString = fmt.Sprintf("%s: %s", turkis.Render(rsp.RspName), rsp.Content)

	return rspString
}

// waitForExternalResponse starts the ResponsePoller() and notifies the Update method
// if a Response comes in automatically
func (m *model) waitForExternalResponse() tea.Cmd {
	return func() tea.Msg {
		return m.userService.ResponsePoller()
	}
}

func (m *model) waitForLog() tea.Cmd {
	return func() tea.Msg {
		return m.LogPoller()
	}
}

func (m *model) LogPoller() t.Log {
	log, ok := <-m.logChan
	if !ok {
		return t.Log{Text: "logging channel is closed", Method: "LogPoller"}
	}

	return log
}

func (m *model) PrintLog(rsp t.Log) {
	m.logs = append(m.logs, fmt.Sprintf("%s: %s", rsp.Method, rsp.Text))
	m.refreshLogViewPort()
}

// ShortHelp decides what to see in the short help window
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Complete, k.Logs, k.NextSug, k.PrevSug}
}

// ShortHelp decides what to see in the extended help window
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.HalfPageUp, k.HalfPageDown, k.InputRight, k.InputLeft},
		{k.Help, k.Quit, k.Complete, k.NextSug, k.PrevSug}}
}
