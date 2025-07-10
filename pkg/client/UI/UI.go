package UI

import (
	"fmt"
	"strings"
	"sync"

	i "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/input"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/types"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

var (
	red    lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#BF3535"))
	blue   lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#353EBF"))
	turkis lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#35BFBC"))
	green  lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("##53BF35"))
)

type (
	errMsg error
)

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	outputChan  chan *t.Response
	userService *i.UserService
	err         error
	mu          *sync.RWMutex
	typing      string
}

func InitialModel(u *i.UserService) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.KeyMap = initialiseKeyMap()

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		err:         nil,
		userService: u,
		outputChan:  u.ChatClient.Output,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.waitForExternalResponse())
}

// TODO chat hochscollen mit Pfeiltasten oder Mausrad und help Fenster
// -> dafür vordefinierte keybinds (zB j & k & u usw) umschreiben
// TODO Suggestions
func (m model) Update(rsp tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		tiCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(rsp)
	m.viewport, vpCmd = m.viewport.Update(rsp)

	m.typing = fmt.Sprint(m.textarea.Value())

	switch rsp := rsp.(type) {
	case *t.Response:
		m.HandleResponse(rsp)
		m.textarea.InsertString(m.typing)
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

	case errMsg:
		m.err = rsp
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"Welcome to the chat room! \nRegister yourself with '/register {name}'\n%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}

func (m *model) HandleWindowResize(rsp *tea.WindowSizeMsg) {
	m.viewport.Width = rsp.Width
	m.textarea.SetWidth(rsp.Width)
	m.viewport.Height = rsp.Height - m.textarea.Height() - lipgloss.Height(gap)

	if len(m.messages) > 0 {
		// Wrap content before setting it. -> Zeilenumbruch am Rand des Viewportes
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
	m.textarea.Reset()
	m.viewport.GotoBottom()
}

func (m *model) Execute() {
	str, _ := strings.CutSuffix(strings.Join(m.messages, "\n"), "\n")
	m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(str))
	m.userService.Executor(m.textarea.Value())
	m.textarea.Reset()
	m.typing = ""
}

func (m *model) evaluateReponse(rsp *t.Response) string {
	var rspString string

	if rsp.Err != nil {
		return red.Render(rsp.Err.Error())
	}

	if rsp.Content == "" {
		return ""
	}

	if strings.HasPrefix(rsp.Content, "[") {
		output, err := JSONToTable(rsp.Content)
		if err != nil {
			return red.Render(fmt.Sprintf("%v: error formatting json to table", err))
		}

		return output
	}

	if rsp.Name == "" {
		rspString = fmt.Sprintf("%s", blue.Render(rsp.Content))
		return rspString
	}

	rspString = fmt.Sprintf("%s: %s", turkis.Render(rsp.Name), rsp.Content)

	return rspString
}

func (m *model) waitForExternalResponse() tea.Cmd {
	return func() tea.Msg {
		return m.userService.ResponsePoller()
	}
}

func initialiseKeyMap() viewport.KeyMap {
	return viewport.KeyMap{
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+down"),
			key.WithHelp("Ctrl+↓", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+up"),
			key.WithHelp("Ctrl+↑", "page up"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("shift+up"),
			key.WithHelp("Shift+↑", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("shift+down"),
			key.WithHelp("Shift+↓", "½ page down"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "right"),
		),
	}
}

// // DisplayResponse prints out a Response in the proper way
// func DisplayResponse(rsp *t.Response) error {
// 	if rsp.Content == "" {
// 		return nil
// 	}

// 	if strings.HasPrefix(rsp.Content, "[") {
// 		output, err := JSONToTable(rsp.Content)
// 		if err != nil {
// 			return fmt.Errorf("%w: error formatting json to table", err)
// 		}

// 		fmt.Println(output)

// 		return nil
// 	}

// 	responseString := fmt.Sprintf("%s: %s", rsp.Name, rsp.Content)
// 	fmt.Println(responseString)

// 	return nil
// }
