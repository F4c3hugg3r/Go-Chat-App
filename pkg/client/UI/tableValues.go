package UI

import (
	"fmt"
	"sync"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type Table struct {
	cols    []table.Column
	rows    []table.Row
	focused bool
	height  int
	width   int
	clients map[string]*ty.JsonClient
	ts      table.Styles
	mu      *sync.RWMutex
}

func setUpTable() (table.Model, *Table) {
	tV := &Table{
		cols: []table.Column{
			{Title: "Name", Width: 5},
			{Title: "Call", Width: 5},
			{Title: "Group", Width: 5},
			{Title: "ClientId", Width: 0},
			{Title: "GroupId", Width: 0},
		},
		rows:    []table.Row{},
		focused: false,
		height:  5,
		clients: make(map[string]*ty.JsonClient),
		mu:      &sync.RWMutex{},
	}

	ta := table.New(
		table.WithColumns(tV.cols),
		table.WithRows(tV.rows),
		table.WithHeight(tV.height),
		table.WithFocused(tV.focused),
	)

	ts := table.DefaultStyles()
	ts.Header = ts.Header.
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(purple.GetForeground()).
		Bold(false)
	ts.Selected = ts.Selected.
		Foreground(lipgloss.NoColor{}).
		Background(lipgloss.NoColor{}).
		Bold(false)
	ta.SetStyles(ts)
	tV.ts = ts

	return ta, tV
}

func (t *Table) SetClients(clients []*ty.JsonClient, client *ty.JsonClient) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if client != nil {
		t.clients[client.ClientId] = client
		return
	}

	clear(t.clients)
	for _, client := range clients {
		t.clients[client.ClientId] = client
	}
}

func (t *Table) ConvertClientsToRows() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rows = t.rows[:0]

	for _, client := range t.clients {
		t.rows = append(t.rows, []string{
			client.Name,
			client.CallState,
			client.GroupName,
			client.ClientId,
			client.GroupId,
		})
	}
}

func (t *Table) SetCallState(clientId string, callState string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	client, exists := t.clients[clientId]
	if !exists {
		return fmt.Errorf("%w: client is not in table", ty.ErrNotAvailable)
	}

	client.CallState = callState
	return nil
}

func (t *Table) AddClient(client *ty.JsonClient) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.clients[client.ClientId]; exists {
		return fmt.Errorf("%w: client already exists", ty.ErrNoPermission)
	}

	t.clients[client.ClientId] = client
	return nil
}

func (t *Table) RemoveClient(clientId string, all bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if all {
		clear(t.clients)
		return nil
	}

	if _, exists := t.clients[clientId]; !exists {
		return fmt.Errorf("%w: client does not exist", ty.ErrNotAvailable)
	}

	delete(t.clients, clientId)
	return nil
}

func (t *Table) ChangeSelectedStyle() {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch t.focused {
	case true:
		t.ts.Selected = t.ts.Selected.
			Foreground(lipgloss.NoColor{}).
			Background(lipgloss.NoColor{}).
			Bold(false)
	case false:
		t.ts.Selected = t.ts.Selected.
			Background(purple.GetForeground()).
			Bold(false)
	}
	t.focused = !t.focused
}

func (t *Table) Empty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.clients) > 0 {
		return false
	}

	return true
}
