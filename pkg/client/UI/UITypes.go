package UI

import (
	i "github.com/F4c3hugg3r/Go-Chat-Server/pkg/client/input"
	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

const Gap = "\n\n"
const RegisterTitle = "Du bist registriert %s!"
const UnregisterTitle = "Willkommen im Chatraum! \nSchreibe '/register {name}' und '/help'"
const GroupTitle = "%s, du bist in der Gruppe %s!"
const WindowResizeFlag = "windowResize"
const RegisterOutput = "-> Du kannst nun Nachrichten schreiben oder Commands ausführen" +
	"\n		'/help' → Befehle anzeigen" +
	"\n		'/quit' → Chat verlassen" +
	"\n		'/users' → Tabelle aktualisieren"

var (
	red    lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#BF3535"))
	redBg  lipgloss.Style = lipgloss.NewStyle().Background(lipgloss.Color("#BF3535"))
	noCol  lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.NoColor{})
	blue   lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#3571bfff"))
	purple lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	turkis lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#35BFBC"))
	// green  lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#53BF35"))

	titleStyle lipgloss.Style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			PaddingLeft(10).
			PaddingRight(10)
	TableHeaderStyle = table.DefaultStyles().Header.
				BorderBottom(true).
				Bold(false)
	TableSelectedStyle = table.DefaultStyles().Selected.
				Background(lipgloss.Color("64")).
				Bold(false)
	centered lipgloss.Style = lipgloss.NewStyle().Align(lipgloss.Center)
	faint    lipgloss.Style = lipgloss.NewStyle().Faint(true)

	viewportKeys = viewport.KeyMap{
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
		SelectUser:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl u", faint.Render("select user"))),
		Logs:         key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl l", faint.Render("toggle logs"))),
		MuteMic:      key.NewBinding(key.WithKeys("alt+m"), key.WithHelp("alt m", faint.Render("mute microphone"))),
		MuteSpeaker:  key.NewBinding(key.WithKeys("alt+s"), key.WithHelp("alt s", faint.Render("mute speaker & mic"))),
		NextSug:      key.NewBinding(key.WithKeys("shift+right"), key.WithHelp("shift →", faint.Render("next suggestion"))),
		PrevSug:      key.NewBinding(key.WithKeys("shift+left"), key.WithHelp("shift ←", faint.Render("previous suggestion"))),
		HalfPageUp:   viewportKeys.HalfPageUp,
		HalfPageDown: viewportKeys.HalfPageDown,
		Down:         viewportKeys.Down,
		Up:           viewportKeys.Up,
		InputLeft:    key.NewBinding(key.WithKeys("alt+c"), key.WithHelp("alt c", faint.Render("left previous input"))),
		InputRight:   key.NewBinding(key.WithKeys("alt+v"), key.WithHelp("alt v", faint.Render("right previous input"))),
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
	Logs         key.Binding
	SelectUser   key.Binding
	Complete     key.Binding
	NextSug      key.Binding
	PrevSug      key.Binding
	MuteMic      key.Binding
	MuteSpeaker  key.Binding
}

// model contains every view-model and variables/structs needed for
// the displaying and handling of the TUI
type model struct {
	viewport viewport.Model
	// logging
	logViewport     viewport.Model
	logs            []string
	messages        []string
	logChan         chan t.Log
	title           string
	textinput       textinput.Model
	showSuggestions bool
	helpModel       help.Model
	keyMap          keyMap
	err             error

	table           table.Model
	tableValues     *Table
	muteTable       table.Model
	muteTableValues *MuteTable

	userService *i.UserService
	outputChan  chan *t.Response

	inH *InputHistory
}

// InputHistory manageges the inputHistory
type InputHistory struct {
	current int
	inputs  []string
	first   bool
}
