package UI

import (
	"fmt"
	"strings"
	"time"

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

// TODO call accept / deny functionality

// maybe TODO bg farben für connection und mute -> mute in server speichern und übergeben -> rot faint => mic, rot => mic + speaker

// InitialModel initializes the model struct, which is the main struct for the TUI
func InitialModel(u *i.UserService) model {
	ti := SetUpTextInput(u)

	h := help.New()

	ta, tV := setUpTable(u.Client.LogChan)
	mTa, mTV := setUpMuteTable(u.Client.LogChan)

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
		outputChan:      u.Client.Output,
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

	model.logChan = model.userService.Client.LogChan

	return model
}

// Init is being called before Update listenes and initializes required functions
func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.waitForExternalResponse(), m.waitForLog(), m.waitForClientsChangeSignal())
}

// Update handles every input
func (m model) Update(rsp tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd  tea.Cmd
		tiCmd  tea.Cmd
		loCmd  tea.Cmd
		tbCmd  tea.Cmd
		mTbCmd tea.Cmd
	)

	m.viewport, vpCmd = m.viewport.Update(rsp)
	m.textinput, tiCmd = m.textinput.Update(rsp)
	m.logViewport, loCmd = m.logViewport.Update(rsp)
	m.table, tbCmd = m.table.Update(rsp)
	m.muteTable, mTbCmd = m.muteTable.Update(rsp)

	if m.textinput.Value() == "" {
		m.inH.SaveInput("")
	}

	switch rsp := rsp.(type) {
	case CallResultMsg:
		if rsp.Accepted {
			m.DisplayMessage(green.Render("- Du hast den Anruf angenommen -"))
		} else {
			m.DisplayMessage(green.Render("- Du hast den Anruf abgelehnt -"))
		}
		m.refreshTable("")

	case t.ClientsChangeSignal:
		m.HandleClientsChangeSignal(rsp)

		return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd, mTbCmd, m.waitForClientsChangeSignal())

	case t.Log:
		m.PrintLog(rsp)

		return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd, mTbCmd, m.waitForLog())

	case *t.Response:
		m.HandleResponse(rsp)

		if strings.Contains(rsp.Content, t.ReceiveCall) {
			return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd, mTbCmd, m.ReceiveCall(rsp), m.waitForExternalResponse())
		}

		return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd, mTbCmd, m.waitForExternalResponse())

	case tea.WindowSizeMsg:
		m.HandleWindowResize(&rsp)
		m.logChan <- t.Log{Text: "log started"}

	case tea.KeyMsg:
		switch rsp.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.userService.Client.Interrupt()

			return m, tea.Quit

		case tea.KeyEnter:
			if m.table.Focused() {
				// maybe TODO ausgewähltes Feld in zwischenablage oder in textinput
				// mit zB zuvor eingegebenen Text kopieren oder automatisch suggesten
				m.HandleTableSelect()
				return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd, mTbCmd)
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

		case key.Matches(rsp, m.keyMap.MuteMic):
			m.HandleMute(t.Microphone)

		case key.Matches(rsp, m.keyMap.MuteSpeaker):
			m.HandleMute(t.Speaker)

		case key.Matches(rsp, m.keyMap.SelectUser):
			m.HandleTableFocus()

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

	return m, tea.Batch(tiCmd, vpCmd, loCmd, tbCmd, mTbCmd)
}

// View describes the terminal view
func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s%s%s%s%s%s%s",
		m.title,
		Gap,
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			m.viewport.View(),
			m.table.View()),
		Gap,
		m.textinput.View(),
		Gap,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(m.helpModel.Width).Render(m.helpModel.View(m.keyMap)),
			lipgloss.NewStyle().Width(m.muteTable.Width()+2*m.muteTableValues.GetFrameSize()).Render(m.muteTable.View()),
		),
		"\n",
		m.logViewport.View(),
	)
}

// logic functions
func (m *model) ReceiveCall(rsp *t.Response) tea.Cmd {
	m.userService.Client.SetCurrentCalling(rsp.ClientId)
	m.refreshTable(rsp.ClientId)
	m.logChan <- t.Log{Text: fmt.Sprintf("ReceiveCall started for ClientId: %s", rsp.ClientId), Method: "ReceiveCall"}

	return func() tea.Msg {
		select {
		case <-time.After(15 * time.Second):
			message := m.userService.ParseInputToMessage("/call deny")
			m.logChan <- t.Log{Text: "Call timed out after 15 seconds, denying call", Method: "ReceiveCall"}
			m.userService.Client.AnswerCallInitialization(message, t.CallDenied)
			return CallResultMsg{Accepted: false}
		case accepted := <-m.userService.Client.CallTimeoutChan:
			m.logChan <- t.Log{Text: fmt.Sprintf("CallTimeoutChan received: %v", accepted), Method: "ReceiveCall"}
			return CallResultMsg{Accepted: accepted}
		}
	}
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

func (m *model) ToggleMuteTable() {
	switch m.muteTable.Height() {
	case 0:
		m.muteTable.SetHeight(2)
		m.muteTable.SetColumns([]table.Column{})
		// m.muteTable.SetRows(m.muteTableValues.rows)
		m.textinput.Width = m.textinput.Width - m.muteTable.Width()
	case 2:
		m.muteTable.SetHeight(0)
		m.muteTable.SetColumns(m.muteTableValues.cols)
		// m.muteTable.SetRows([]table.Row{})
		m.textinput.Width = m.textinput.Width + m.muteTable.Width()
	}
}

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
	m.refreshTable("")
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

// RenderTitle decides between the registered and unregistered title sets it into the viewport
// and return the heightDiff of the old and new title
func (m *model) RenderTitle(title string, param []string) {
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

func (m *model) refreshTable(clientToBlink string) {
	m.tableValues.ConvertClientsToRows(clientToBlink)
	m.table.SetRows(m.tableValues.rows)
	m.table.SetStyles(m.tableValues.ts)
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

func (m *model) HandleTableFocus() {
	if m.userService.Client.GetGroupId() != "" {
		m.userService.Executor("/group users")
	} else {
		m.userService.Executor("/users")
	}
	if m.tableValues.Empty() {
		m.DisplayMessage(red.Render(fmt.Sprintf("%v: register yourself first", t.ErrNoPermission)))
		return
	}
	m.SwitchFocus()
}

func (m *model) HandleMute(toMute string) {
	if !m.userService.Client.Registered {
		m.DisplayMessage(red.Render("you are not registered yet"))
		return
	}

	err := m.userService.Client.Mute(toMute)
	if err != nil {
		m.DisplayMessage(red.Render("you have to be in a call first"))
		return
	}
	m.muteTableValues.SetMute(toMute)
	m.muteTable.SetColumns(m.muteTableValues.cols)
}

func (m *model) HandleTableSelect() {
	message := turkis.BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).BorderForeground(purple.GetForeground()).Render(
		fmt.Sprintf("%s%s\n%s%s\n%s%s",
			blue.Render("Name:		"),
			turkis.Bold(true).Render(fmt.Sprintf("%s", m.table.SelectedRow()[0])),
			blue.Render("ClientId:	"),
			m.table.SelectedRow()[3],
			blue.Render("GroupId:	 "),
			m.table.SelectedRow()[4],
		))
	m.DisplayMessage(message)
	m.SwitchFocus()

}

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

	m.refreshTable("")
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

	m.helpModel.Width = rsp.Width/8*7 - m.muteTableValues.GetFrameSize()*2

	// m.logChan <- t.Log{Text: fmt.Sprintf("rsp.Width = %d", rsp.Width)}

	m.muteTable.SetWidth(rsp.Width / 8)
	m.muteTable.Columns()[0].Width = m.muteTable.Width() / 2
	m.muteTable.Columns()[1].Width = m.muteTable.Width() / 2

	// m.logChan <- t.Log{Text: fmt.Sprintf("helpWidth = %d, viewport width = %d", m.helpModel.Width, m.viewport.Width)}
	// m.logChan <- t.Log{Text: fmt.Sprintf("muteTableWidth = %d", m.muteTable.Width())}

	m.textinput.Width = rsp.Width
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

	m.RenderTitle(m.title, []string{WindowResizeFlag})
	m.refreshViewPort()
	m.refreshTable("")
}

// HandleResponse handles an incoming Response by evaluating it and refreshing
// the viewport by adding the corresponding string
func (m *model) HandleResponse(rsp *t.Response) {
	str := m.EvaluateReponse(rsp)
	if str != "" {
		m.DisplayMessage(str)
	}

	// m.refreshViewPort()
}

// HandleMessage hanbles the input message by saving it for the inputHistory and
// executeing the fitting userService method
func (m *model) HandleMessage() {
	m.inH.SaveInput(m.textinput.Value())

	m.userService.Executor(m.textinput.Value())
	m.textinput.Reset()
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
		return m.clientsChangeSignalPoller()
	}
}

func (m *model) clientsChangeSignalPoller() t.ClientsChangeSignal {
	signal, ok := <-m.userService.Client.ClientChangeSignalChan
	if !ok {
		m.logChan <- t.Log{Text: "ClientsChangeSignal channel is closed"}
		return t.ClientsChangeSignal{}
	}

	return signal
}

//
// helper functions
//

func (m *model) DisplayMessage(message string) {
	m.messages = append(m.messages, message)
	m.refreshViewPort()
}

func (m *model) PrintLog(rsp t.Log) {
	m.logs = append(m.logs, fmt.Sprintf("%s: %s", rsp.Method, rsp.Text))
	m.refreshLogViewPort()
}

// ShortHelp decides what to see in the short help window
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Complete, k.SelectUser, k.Quit, k.Logs}
}

// ShortHelp decides what to see in the extended help window
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.HalfPageUp, k.HalfPageDown, k.NextSug, k.PrevSug},
		{k.Help, k.SelectUser, k.Logs, k.Quit, k.Complete},
		{k.MuteMic, k.MuteSpeaker, k.InputRight, k.InputLeft},
	}
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
