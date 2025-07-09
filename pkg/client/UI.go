package client

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

type (
	errMsg error
)

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	outputChan  chan *Response
	userService *UserService
	err         error
}

func InitialModel(u *UserService) model {
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
	vp.SetContent(`Welcome to the chat room! Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
		userService: u,
		outputChan:  u.chatClient.Output,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.waitForExternalResponse())
}

func (m model) Update(rsp tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(rsp)
	m.viewport, vpCmd = m.viewport.Update(rsp)

	switch rsp := rsp.(type) {
	case *Response:
		m.messages = append(m.messages, m.evaluateReponse(rsp))
		// m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		m.textarea.Reset()
		m.viewport.GotoBottom()
		return m, tea.Batch(tiCmd, vpCmd, m.waitForExternalResponse())
	// TODO chat hochscollen mit Pfeiltasten oder Mausrad und help Fenster

	// for terminal resize
	case tea.WindowSizeMsg:
		m.viewport.Width = rsp.Width
		m.textarea.SetWidth(rsp.Width)
		m.viewport.Height = rsp.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch rsp.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
			m.messages = append(m.messages, m.textarea.Value())
			m.messages = append(m.messages, rsp.String())
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
			m.userService.Executor(m.textarea.Value())
			m.textarea.Reset()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = rsp
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}

func (m *model) evaluateReponse(rsp *Response) string {
	var rspString string

	if rsp.Content == "" {
		return rsp.Err.Error()
	}

	if rsp.Err != nil {
		// TODO Farbe
		return rsp.Err.Error()
	}

	if strings.HasPrefix(rsp.Content, "[") {
		output, err := JSONToTable(rsp.Content)
		if err != nil {
			return fmt.Sprintf("%v: error formatting json to table", err)
		}

		return output
	}

	// TODO Farbe
	//zB m.senderStyle.Render
	rspString = fmt.Sprintf("%s: %s", rsp.Name, rsp.Content)

	return rspString
}

func (m *model) waitForExternalResponse() tea.Cmd {
	return func() tea.Msg {
		return m.userService.ResponsePoller()
	}
}

// DisplayResponse prints out a Response in the proper way
func DisplayResponse(rsp *Response) error {
	if rsp.Content == "" {
		return nil
	}

	if strings.HasPrefix(rsp.Content, "[") {
		output, err := JSONToTable(rsp.Content)
		if err != nil {
			return fmt.Errorf("%w: error formatting json to table", err)
		}

		fmt.Println(output)

		return nil
	}

	responseString := fmt.Sprintf("%s: %s", rsp.Name, rsp.Content)
	fmt.Println(responseString)

	return nil
}
