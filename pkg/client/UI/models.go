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
const registerflag = "- Du bist registriert -"
const unregisterFlag = "- Du bist nun vom Server getrennt -"

var (
	red          lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#BF3535"))
	blue         lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#353EBF"))
	purple       lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	turkis       lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#35BFBC"))
	green        lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("##53BF35"))
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
		// Left: key.NewBinding(
		// 	key.WithKeys("left"),
		// 	key.WithHelp("←", faint.Render("move left")),
		// ),
		// Right: key.NewBinding(
		// 	key.WithKeys("right"),
		// 	key.WithHelp("→", faint.Render("move right")),
		// ),
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
		// Left:         viewportKeys.Left,
		// Right:        viewportKeys.Right,
	}
)

type (
	errMsg error
)

type keyMap struct {
	Up   key.Binding
	Down key.Binding
	// Left         key.Binding
	// Right        key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	Help         key.Binding
	Quit         key.Binding
	Complete     key.Binding
	NextSug      key.Binding
	PrevSug      key.Binding
}

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
}
