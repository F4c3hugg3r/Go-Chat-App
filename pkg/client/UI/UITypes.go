package UI

import (
	"encoding/json"
	"strings"

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
const UnregisterTitle = "Willkommen im Chatraum! \nSchreibe '/register {name}' oder '/help'"
const GroupTitle = "%s, du bist in der Gruppe %s!"
const WindowResizeFlag = "windowResize"

var (
	red    lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#BF3535"))
	blue   lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#3571bfff"))
	purple lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	turkis lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#35BFBC"))
	// green  lipgloss.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#53BF35"))

	titleStyle lipgloss.Style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			PaddingLeft(5).
			PaddingRight(5)
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
		Logs:         key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl l", faint.Render("toggle logs"))),
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
	Logs         key.Binding
	Complete     key.Binding
	NextSug      key.Binding
	PrevSug      key.Binding
}

// model contains every view-model and variables/structs needed for
// the displaying and handling of the TUI
type model struct {
	viewport viewport.Model
	// logging
	logViewport     viewport.Model
	table           table.Model
	logs            []string
	messages        []string
	logChan         chan t.Log
	title           string
	textinput       textinput.Model
	showSuggestions bool
	help            help.Model
	keyMap          keyMap
	tableValues     *Table
	err             error

	userService *i.UserService
	outputChan  chan *t.Response

	inH *InputHistory
}

type UIClient struct {
	Name      string
	CallState string
	ClientId  string
	Active    bool
}

// InputHistory manageges the inputHistory
type InputHistory struct {
	current int
	inputs  []string
	first   bool
}

// helper functions
func ParseJsonToUIClients(jsonSlice string) ([]*UIClient, error) {
	var clients []*UIClient
	dec := json.NewDecoder(strings.NewReader(jsonSlice))

	err := dec.Decode(&clients)
	if err != nil {
		return nil, err
	}

	return clients, nil
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
