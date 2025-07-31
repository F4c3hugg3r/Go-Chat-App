package UI

import (
	"sync"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type Table struct {
	cols    []table.Column
	rows    []table.Row
	focused bool
	height  int
	width   int
	clients map[string]*UIClient
	mu      *sync.RWMutex
}

func setUpTable() (table.Model, *Table) {
	tV := &Table{
		cols: []table.Column{
			{Title: "Name", Width: 5},
			{Title: "Call", Width: 5},
		},
		rows:    []table.Row{},
		focused: false,
		height:  5,
		clients: make(map[string]*UIClient),
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
		Background(purple.GetForeground()).
		Bold(false)
	ta.SetStyles(ts)

	return ta, tV
}

func (t *Table) SetClients(clients []*UIClient, client *UIClient) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if client != nil {
		t.clients[client.ClientId] = client
		return
	}

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
			client.Name, client.CallState,
		})
	}
}
