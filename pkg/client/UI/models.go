package UI

import (
	i "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/input"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/types"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"
const registerFlag = "- Du bist registriert -"
const registerTitle = "Du bist registriert %s!"
const unregisterFlag = "- Du bist nun vom Server getrennt -"
const unregisterTitle = "Willkommen im Chatraum! \nSchreibe '/register {name}' oder '/help'"
const addGroupFlag = "Add Group"
const leaveGroupFlag = "Leave Group"
const GroupTitle = "%s, du bist in der Gruppe %s!"

var (
	red        lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#BF3535"))
	blue       lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#3571bfff"))
	purple     lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	turkis     lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#35BFBC"))
	green      lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("##53BF35"))
	titleStyle lipgloss.Style = lipgloss.NewStyle().
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			PaddingLeft(5).
			PaddingRight(5)
	centered     lipgloss.Style = lipgloss.NewStyle().Align(lipgloss.Center)
	faint        lipgloss.Style = lipgloss.NewStyle().Faint(true)
	viewportKeys                = viewport.KeyMap{
		HalfPageUp: key.NewBinding(
			key.WithKeys("shift+up"),
			key.WithHelp("shift ↑", faint.Render("move ½ page up")),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("shift+down"),
			key.WithHelp("shift ↓", faint.Render("move ½ page down")),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", faint.Render("move down")),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", faint.Render("move up")),
		),
	}
	helpKeys = keyMap{
		Help:         key.NewBinding(key.WithKeys("ctrl+h"), key.WithHelp("ctrl h", faint.Render("toggle help"))),
		Complete:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", faint.Render("complete"))),
		Quit:         key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", faint.Render("quit"))),
		NextSug:      key.NewBinding(key.WithKeys("shift+right"), key.WithHelp("shift →", faint.Render("next suggestion"))),
		PrevSug:      key.NewBinding(key.WithKeys("shift+left"), key.WithHelp("shift ←", faint.Render("previous suggestion"))),
		HalfPageUp:   viewportKeys.HalfPageUp,
		HalfPageDown: viewportKeys.HalfPageDown,
		Down:         viewportKeys.Down,
		Up:           viewportKeys.Up,
		InputLeft:    key.NewBinding(key.WithKeys("alt+n"), key.WithHelp("alt n", faint.Render("left previous input"))),
		InputRight:   key.NewBinding(key.WithKeys("alt+m"), key.WithHelp("alt m", faint.Render("right previous input"))),
	}
)

type (
	errMsg error
)

// keyMap contains the keybinds for the TUI
type keyMap struct {
	Up           key.Binding
	Down         key.Binding
	InputLeft    key.Binding
	InputRight   key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	Help         key.Binding
	Quit         key.Binding
	Complete     key.Binding
	NextSug      key.Binding
	PrevSug      key.Binding
}

// model contains every view-model and variables/structs needed for
// the displaying and handling of the TUI
type model struct {
	viewport        viewport.Model
	textinput       textinput.Model
	help            help.Model
	messages        []string
	outputChan      chan *t.Response
	userService     *i.UserService
	err             error
	keyMap          keyMap
	showSuggestions bool
	registered      string
	title           string
	inH             *InputHistory
}

// InputHistory manageges the inputHistory
type InputHistory struct {
	current int
	inputs  []string
	first   bool
}
