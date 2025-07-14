package UI

import (
	"fmt"
	"strings"

	i "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/input"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/types"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// (TODO list commands output als Liste darstellen)

// InitialModel initializes the model struct, which is the main struct for the TUI
func InitialModel(u *i.UserService) model {
	ti := setUpTextInput(u)

	h := help.New()

	vp := viewport.New(30, 5)
	vp.KeyMap = viewportKeys

	inputManager := &InputHistory{
		current: -1,
		inputs:  make([]string, 0, 200),
	}

	model := model{
		messages:    []string{},
		viewport:    vp,
		err:         nil,
		userService: u,
		outputChan:  u.ChatClient.Output,
		textinput:   ti,
		help:        h,
		keyMap:      helpKeys,
		inH:         inputManager,
		registered:  unregisterFlag,
	}

	return model
}

// Init is being called before Update listenes and initializes required functions
func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.waitForExternalResponse())
}

// Update handles every input
func (m model) Update(rsp tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		tiCmd tea.Cmd
	)

	m.viewport, vpCmd = m.viewport.Update(rsp)
	m.textinput, tiCmd = m.textinput.Update(rsp)

	if m.textinput.Value() == "" {
		m.inH.SaveInput("")
	}

	switch rsp := rsp.(type) {
	case *t.Response:
		m.HandleResponse(rsp)

		return m, tea.Batch(tiCmd, vpCmd, m.waitForExternalResponse())

	case tea.WindowSizeMsg:
		m.HandleWindowResize(&rsp)

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

		case key.Matches(rsp, m.keyMap.InputLeft), key.Matches(rsp, m.keyMap.InputRight):
			input := m.SearchInputHistory(rsp)
			m.textinput.SetValue(input)
			m.textinput.CursorEnd()
		}

	case errMsg:
		m.err = rsp
		return m, nil
	}

	m.setTitle()

	return m, tea.Batch(tiCmd, vpCmd, tiCmd)
}

// View describes the terminal view
func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s%s%s%s%s",
		m.title,
		gap,
		m.viewport.View(),
		gap,
		m.textinput.View(),
		gap,
		m.help.View(m.keyMap),
	)
}

// refreshViewPort refreshes the size of the viewport
func (m *model) refreshViewPort() {
	if len(m.messages) > 0 {
		str, _ := strings.CutSuffix(strings.Join(m.messages, "\n"), "\n")
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(str))
	}

	m.viewport.GotoBottom()
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

// setUpTexInput sets up a textinput.Model with every needed setting
func setUpTextInput(u *i.UserService) textinput.Model {
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

// func (m *model) refreshTitle(title string) {
// 	heigtDiff := lipgloss.Height(title) - lipgloss.Height(m.title)
// 	m.title = registerTitle
// 	m.viewport.Height = m.viewport.Height - heigtDiff
// }

// setTitle decides between the registered and unregistered title sets it into the viewport
// and return the heightDiff of the old and new title
func (m *model) setTitle() {
	if strings.Contains(m.registered, unregisterFlag) && !strings.Contains(m.title, unregisterTitle) {
		m.title = centered.Width(m.viewport.Width).Bold(true).Render(unregisterTitle)

		return
	}

	if strings.Contains(m.registered, registerFlag) && !strings.Contains(m.title, registerTitle) {
		title := centered.Width(m.viewport.Width).
			Render(titleStyle.Render(fmt.Sprintf(registerTitle,
				turkis.Render(m.userService.ChatClient.GetName()))))

		heigtDiff := lipgloss.Height(title) - lipgloss.Height(m.title)
		m.title = registerTitle
		m.viewport.Height = m.viewport.Height - heigtDiff
	}
}

// HandleWindowResize handles rezising of the terminal window by updating all models sizes
func (m *model) HandleWindowResize(rsp *tea.WindowSizeMsg) {
	m.viewport.Width = rsp.Width
	m.textinput.Width = rsp.Width
	m.help.Width = rsp.Width
	m.viewport.Height = rsp.Height - lipgloss.Height(gap) - lipgloss.Height(m.title) - lipgloss.Height(gap)

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

// evaluateResponse evaluates an incoming Response and returns the
// corresponding rendered string
func (m *model) evaluateReponse(rsp *t.Response) string {
	var rspString string

	switch {
	// error output
	case rsp.Err != nil:
		return red.Render(rsp.Err.Error())

	// empty output
	case rsp.Content == "":
		return ""

	// slice output
	case strings.HasPrefix(rsp.Content, "["):
		output, err := JSONToTable(rsp.Content)
		if err != nil {

			return red.Render(fmt.Sprintf("%v: error formatting json to table", err))
		}

		return output

	// register output
	case strings.Contains(rsp.Content, registerFlag):
		m.registered = rsp.Content

		return blue.Render("-> Du kannst nun Nachrichten schreiben oder Commands ausführen\n'/help' → Befehle anzeigen\n'/quit' → Chat verlassen")

	// server output
	case rsp.Name == "":
		rspString = fmt.Sprintf("%s", blue.Render(rsp.Content))

		// 	unregister output
		if strings.Contains(rsp.Content, unregisterFlag) {
			m.registered = unregisterFlag
		}

		return rspString
	}

	// response output
	rspString = fmt.Sprintf("%s: %s", turkis.Render(rsp.Name), rsp.Content)

	return rspString
}

// waitForExternalResponse starts the ResponsePoller() and notifies the Update method
// if a Response comes in automatically
func (m *model) waitForExternalResponse() tea.Cmd {
	return func() tea.Msg {
		return m.userService.ResponsePoller()
	}
}

// ShortHelp decides what to see in the short help window
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Complete, k.NextSug, k.PrevSug}
}

// ShortHelp decides what to see in the extended help window
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.HalfPageUp, k.HalfPageDown, k.InputRight, k.InputLeft},
		{k.Help, k.Quit, k.Complete, k.NextSug, k.PrevSug}}
}
