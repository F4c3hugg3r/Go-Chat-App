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

// TODO list commands output als Liste darstellen

// TODO Ordnerstruktur Server
// TODO eigenen Namen anzeigen
// TODO shortcut für last input (Pfeiltasten)

func InitialModel(u *i.UserService) model {
	ti := setUpTextInput(u)

	h := help.New()

	vp := viewport.New(30, 5)
	vp.KeyMap = viewportKeys

	return model{
		messages:    []string{},
		viewport:    vp,
		err:         nil,
		userService: u,
		outputChan:  u.ChatClient.Output,
		textinput:   ti,
		help:        h,
		keyMap:      helpKeys,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.waitForExternalResponse())
}

func (m model) Update(rsp tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		tiCmd tea.Cmd
	)

	m.viewport, vpCmd = m.viewport.Update(rsp)
	m.textinput, tiCmd = m.textinput.Update(rsp)

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
			m.Execute()
		}

		if key.Matches(rsp, m.keyMap.Help) {
			m.help.ShowAll = !m.help.ShowAll
		}

	case errMsg:
		m.err = rsp
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd, tiCmd)
}

func (m model) View() string {
	m.setTitle()

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

func (m *model) setTitle() {
	if m.registered == "" {
		m.title = centered.Width(m.viewport.Width).Bold(true).Render("Welcome to the chat room! \nTry '/register {name}' or '/help'")
		return
	}

	title := centered.Width(m.viewport.Width).Bold(true).Render(m.registered)
	m.title = title
	heigtDiff := lipgloss.Height(m.title) - lipgloss.Height(title)
	m.viewport.Height = m.viewport.Height + heigtDiff
}

func (m *model) HandleWindowResize(rsp *tea.WindowSizeMsg) {
	m.viewport.Width = rsp.Width
	m.textinput.Width = rsp.Width
	m.help.Width = rsp.Width
	m.viewport.Height = rsp.Height - lipgloss.Height(gap) - lipgloss.Height(m.title) - lipgloss.Height(gap)

	if len(m.messages) > 0 {
		str, _ := strings.CutSuffix(strings.Join(m.messages, "\n"), "\n")
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(str))
	}
	m.viewport.GotoBottom()
}

func (m *model) HandleResponse(rsp *t.Response) {
	str := m.evaluateReponse(rsp)
	if str != "" {
		m.messages = append(m.messages, str)
	}

	str, _ = strings.CutSuffix(strings.Join(m.messages, "\n"), "\n")
	m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(str))
	m.viewport.GotoBottom()
}

func (m *model) Execute() {
	str, _ := strings.CutSuffix(strings.Join(m.messages, "\n"), "\n")
	m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(str))
	m.userService.Executor(m.textinput.Value())
	m.textinput.Reset()
}

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
	case strings.Contains(rsp.Content, registerflag):
		m.registered = rsp.Content

		return blue.Render("-> Du kannst nun Nachrichten schreiben oder Commands ausführen")

	// server output
	case rsp.Name == "":
		rspString = fmt.Sprintf("%s", blue.Render(rsp.Content))

		// 	unregister output
		if strings.Contains(rsp.Content, unregisterFlag) {
			m.registered = ""
		}

		return rspString
	}

	// response output
	rspString = fmt.Sprintf("%s: %s", turkis.Render(rsp.Name), rsp.Content)

	return rspString
}

func (m *model) waitForExternalResponse() tea.Cmd {
	return func() tea.Msg {
		return m.userService.ResponsePoller()
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Complete, k.NextSug, k.PrevSug}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.HalfPageUp, k.HalfPageDown},
		{k.Help, k.Quit, k.Complete, k.NextSug, k.PrevSug}}
}

// func setUpTextArea(u *i.UserService) textarea.Model {
// 	ta := textarea.New()
// 	ta.Placeholder = "Send a message..."
// 	ta.Focus()
// 	ta.Prompt = "┃ "
// 	ta.CharLimit = 280
// 	ta.SetWidth(30)
// 	ta.SetHeight(3)
// 	// Remove cursor line styling
// 	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
// 	ta.ShowLineNumbers = false
// 	ta.KeyMap.InsertNewline.SetEnabled(false)
// 	return ta
// }
