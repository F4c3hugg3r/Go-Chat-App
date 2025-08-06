package UI

import (
	"fmt"
	"strings"

	i "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/input"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TODO ALLGEMEIN

// TODO table für mute speaker/mic

// TODO mute in server speichern und übergeben -> rot faint => mic, rot => mic + speaker

// InitialModel initializes the model struct, which is the main struct for the TUI
func InitialModel(u *i.UserService) model {
	ti := SetUpTextInput(u)

	h := help.New()

	ta, tV := setUpTable()
	mTa, mTV := setUpMuteTable()

	vp := viewport.New(30, 5)
	vp.KeyMap = viewportKeys

	logVp := viewport.New(30, 0)

	inputManager := &InputHistory{
		current: -1,
		inputs:  make([]string, 0, 200),
	}

	model := model{
		messages:        []string{},
		viewport:        vp,
		logViewport:     logVp,
		logs:            []string{},
		err:             nil,
		userService:     u,
		outputChan:      u.ChatClient.Output,
		textinput:       ti,
		helpModel:       h,
		keyMap:          helpKeys,
		inH:             inputManager,
		title:           UnregisterTitle,
		table:           ta,
		tableValues:     tV,
		muteTable:       mTa,
		muteTableValues: mTV,
	}

	model.logChan = model.userService.ChatClient.LogChan

	return model
}

// Init is being called before Update listenes and initializes required functions
func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.waitForExternalResponse(), m.waitForLog(), m.waitForClientsChangeSignal())
}

// Update handles every input
func (m model) Update(rsp tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		tiCmd tea.Cmd
		loCmd tea.Cmd
		tbCmd tea.Cmd
	)

	m.viewport, vpCmd = m.viewport.Update(rsp)
	m.textinput, tiCmd = m.textinput.Update(rsp)
	m.logViewport, loCmd = m.logViewport.Update(rsp)
	m.table, tbCmd = m.table.Update(rsp)

	if m.textinput.Value() == "" {
		m.inH.SaveInput("")
	}

	switch rsp := rsp.(type) {
	case t.ClientsChangeSignal:
		m.HandleClientsChangeSignal(rsp)

		return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd, m.waitForClientsChangeSignal())

	case t.Log:
		m.PrintLog(rsp)

		return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd, m.waitForLog())

	case *t.Response:
		m.HandleResponse(rsp)

		return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd, m.waitForExternalResponse())

	case tea.WindowSizeMsg:
		m.HandleWindowResize(&rsp)
		m.logChan <- t.Log{Text: "log started"}

	case tea.KeyMsg:
		switch rsp.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.userService.ChatClient.Interrupt()

			return m, tea.Quit

		case tea.KeyEnter:
			if m.table.Focused() {
				// maybe TODO ausgewähltes Feld in zwischenablage oder in textinput
				// mit zB zuvor eingegebenen Text kopieren oder automatisch suggesten
				message := fmt.Sprintf("%s\n%s%s\n%s%s",
					//maybe TODO align center funktioniert nicht
					turkis.Bold(true).AlignHorizontal(lipgloss.Center).Render(fmt.Sprintf("- %s -", m.table.SelectedRow()[0])),
					blue.Render("ClientId: "),
					m.table.SelectedRow()[3],
					blue.Render("GroupId: "),
					m.table.SelectedRow()[4],
				)
				m.AddMessageToViewport(message)
				m.SwitchFocus()

				return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd)
			}
			m.HandleMessage()
		}

		switch {
		case key.Matches(rsp, m.keyMap.Help):
			m.helpModel.ShowAll = !m.helpModel.ShowAll
			switch m.helpModel.ShowAll {
			case true:
				m.refreshViewPortAndTableHeight(6)
			case false:
				m.refreshViewPortAndTableHeight(-6)
			}

		case key.Matches(rsp, m.keyMap.SelectUser):
			if m.tableValues.Empty() {
				m.AddMessageToViewport(red.Render(fmt.Sprintf("%v: register yourself first", t.ErrNoPermission)))
				return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd)
			}
			m.SwitchFocus()

		case key.Matches(rsp, m.keyMap.Logs):
			m.ToggleLogs()

		case key.Matches(rsp, m.keyMap.InputLeft), key.Matches(rsp, m.keyMap.InputRight):
			input := m.SearchInputHistory(rsp)
			m.textinput.SetValue(input)
			m.textinput.CursorEnd()
		}

	case errMsg:
		m.err = rsp
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd)
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
		m.helpModel.View(m.keyMap),
		Gap,
		m.logViewport.View(),
	)
}

//
// logic functions
//

// TODO bei logs (crtl l) keybinds ändern, sodass textinput sich nicht bewegt
func (m *model) SwitchFocus() {
	switch m.textinput.Focused() {
	case false:
		m.textinput.Focus()
		m.table.Blur()
		m.tableValues.ChangeSelectedStyle()
		m.viewport.KeyMap = viewportKeys
		m.table.KeyMap = table.KeyMap{}

	case true:
		m.table.Focus()
		m.tableValues.ChangeSelectedStyle()
		m.textinput.Blur()
		m.viewport.KeyMap = viewport.KeyMap{}
		m.table.KeyMap = table.DefaultKeyMap()
	}
	m.refreshTable()
	m.refreshViewPort()
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
				turkis.Bold(true).Render(param[0])))

		case strings.Contains(title, t.AddGroupFlag):
			title = titleStyle.Render(fmt.Sprintf(GroupTitle, turkis.Bold(true).Render(param[0]), turkis.Bold(true).Render(param[1])))
		}
	}

	heightDiff := lipgloss.Height(title) - lipgloss.Height(m.title)
	m.title = centered.Width(m.viewport.Width + m.table.Width()).Bold(true).Render(title)
	m.refreshViewPortAndTableHeight(heightDiff)
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
		m.userService.Executor("/users")

		return blue.Render("-> Du kannst nun Nachrichten schreiben oder Commands ausführen" +
			"\n'/help' → Befehle anzeigen\n'/quit' → Chat verlassen\n'/users' → Tabelle aktualisieren")

	// server output
	case rsp.RspName == "":
		rspString = fmt.Sprintf("%s", blue.Render(rsp.Content))

		// unregister output
		if strings.Contains(rsp.Content, t.UnregisterFlag) {
			m.renderTitle(t.UnregisterFlag, nil)
			m.userService.ChatClient.DeletePeersSafely("", true)
			m.userService.ChatClient.OnChangeChan <- t.ClientsChangeSignal{
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

		m.renderTitle(t.AddGroupFlag, []string{m.userService.ChatClient.GetName(), group.Name})
		m.userService.Executor("/group users")

		return fmt.Sprintf("%s %s %s\n%s",
			blue.Render("-> Du bist nun Teil der Gruppe"),
			turkis.Render(group.Name),
			blue.Render("und kannst Nachrichten in ihr schreiben"),
			blue.Faint(true).Render("Private Nachrichten kannst du weiterhin außerhalb verschicken"),
		)

	// leaveGroup output
	case strings.Contains(rsp.RspName, t.LeaveGroupFlag):
		m.userService.ChatClient.UnsetGroupId()
		m.renderTitle(t.RegisterFlag, []string{m.userService.ChatClient.GetName()})
		m.userService.ChatClient.DeletePeersSafely("", true)
		m.userService.Executor("/users")

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

	// // FinishedSignal output
	// case strings.Contains(rsp.RspName, t.OfferSignalFinished):
	// 	m.userService.ChatClient.PushIntoFinishedSignalChan(rsp.ClientId, rsp, t.OfferSignalFinished)

	// slice output
	case strings.HasPrefix(rsp.Content, "["):
		if strings.Contains(rsp.RspName, t.UsersFlag) {
			m.userService.ChatClient.OnChangeChan <- t.ClientsChangeSignal{
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

//
// refresh functions
//

func (m *model) refreshViewPortAndTableHeight(heightDiff int) {
	m.viewport.Height = m.viewport.Height - heightDiff
	m.table.SetHeight(m.viewport.Height)
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
	m.table.SetStyles(m.tableValues.ts)

	// maybe TODO wenn callState connected ist, grün machen
}

// refreshViewPort refreshes the size of the viewport
func (m *model) refreshViewPort() {
	if len(m.messages) > 0 {
		str, _ := strings.CutSuffix(strings.Join(m.messages, "\n"), "\n")
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(str))
	}

	m.viewport.GotoBottom()
}

//
// handler functions
//

func (m *model) HandleClientsChangeSignal(rsp t.ClientsChangeSignal) error {
	// maybe TODO
	// if strings.Contains(rsp.CallState, t.UserAddFlag) {
	// 	client, err := t.DecodeStringToJsonClient(rsp.ClientsJson)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = m.tableValues.AddClient(client)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// maybe TODO
	// if strings.Contains(rsp.CallState, t.UserRemoveFlag) {
	// 	err := m.tableValues.RemoveClient(rsp.OppId, false)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	if strings.Contains(rsp.CallState, t.UnregisterFlag) {
		m.tableValues.RemoveClient("", true)
	}

	if rsp.CallState != "" {
		err := m.tableValues.SetCallState(rsp.OppId, rsp.CallState)
		if err != nil {
			return fmt.Errorf("%w: error setting callstate %s on %s", err, rsp.CallState, rsp.OppId)
		}
	}

	if rsp.ClientsJson != "" {
		clients, err := t.ParseJsonToJsonClients(rsp.ClientsJson)
		if err != nil {
			return fmt.Errorf("%w: error formatting json to clients", err)
		}
		m.tableValues.SetClients(clients, nil)
	}

	m.refreshTable()
	return nil
}

// HandleWindowResize handles rezising of the terminal window by updating all models sizes
func (m *model) HandleWindowResize(rsp *tea.WindowSizeMsg) {
	m.viewport.Width = rsp.Width / 6 * 4

	m.table.SetWidth(rsp.Width / 6 * 2)
	columns := m.table.Columns()
	columns[0].Width = m.table.Width() / 3
	columns[1].Width = m.table.Width() / 3
	columns[2].Width = m.table.Width() / 3
	m.table.SetColumns(columns)

	m.textinput.Width = rsp.Width
	m.helpModel.Width = rsp.Width
	m.logViewport.Width = rsp.Width

	var logPortHeight int
	switch m.logViewport.Height {
	case 0:
		logPortHeight = 1
	default:
		logPortHeight = m.logViewport.Height
	}

	m.viewport.Height = rsp.Height - (lipgloss.Height(Gap) * 2) - lipgloss.Height(m.title) - logPortHeight
	m.table.SetHeight(m.viewport.Height)

	m.renderTitle(m.title, []string{WindowResizeFlag})
	m.refreshViewPort()
	m.refreshTable()
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

//
//	poll channels
//

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

func (m *model) waitForClientsChangeSignal() tea.Cmd {
	return func() tea.Msg {
		return m.ClientsChangeSignalPoller()
	}
}

func (m *model) ClientsChangeSignalPoller() t.ClientsChangeSignal {
	signal, ok := <-m.userService.ChatClient.OnChangeChan
	if !ok {
		m.logChan <- t.Log{Text: "ClientsChangeSignal channel is closed"}
		return t.ClientsChangeSignal{}
	}

	return signal
}

//
// helper functions
//

func (m *model) AddMessageToViewport(message string) {
	m.messages = append(m.messages, message)
	m.refreshViewPort()
}

func (m *model) ToggleLogs() {
	switch m.logViewport.Height {
	case 0:
		m.logViewport.Height = 8
		m.refreshViewPortAndTableHeight(8)
		m.refreshLogViewPort()
		m.viewport.KeyMap = viewport.KeyMap{}
		m.table.KeyMap = table.KeyMap{}
	case 8:
		m.logViewport.Height = 0
		m.refreshViewPortAndTableHeight(-8)
		m.viewport.KeyMap = viewportKeys
		m.table.KeyMap = table.DefaultKeyMap()
	}
	m.refreshViewPort()
}

func (m *model) PrintLog(rsp t.Log) {
	m.logs = append(m.logs, fmt.Sprintf("%s: %s", rsp.Method, rsp.Text))
	m.refreshLogViewPort()
}

// ShortHelp decides what to see in the short help window
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Complete, k.SelectUser, k.Logs, k.NextSug, k.PrevSug}
}

// ShortHelp decides what to see in the extended help window
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.HalfPageUp, k.HalfPageDown, k.InputRight, k.InputLeft},
		{k.Help, k.SelectUser, k.Logs, k.Quit, k.Complete, k.NextSug, k.PrevSug}}
}

// setUpTexInput sets up a textinput.Model with every needed setting
func SetUpTextInput(u *i.UserService) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Send a message..."
	ti.Prompt = "┃ "
	ti.PromptStyle, ti.Cursor.Style = purple, purple
	ti.Focus()
	ti.CharLimit = 280
	ti.Width = 30
	ti.ShowSuggestions = true

	//überschreiben, da ctrl+h sonst char löscht
	ti.KeyMap.DeleteCharacterBackward = key.NewBinding(key.WithKeys("backspace"))
	ti.KeyMap.NextSuggestion = helpKeys.NextSug
	ti.KeyMap.PrevSuggestion = helpKeys.PrevSug

	ti.SetSuggestions(u.InitializeSuggestions())

	return ti
}
