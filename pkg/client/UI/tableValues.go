package UI

import (
	"fmt"
	"sync"

	t "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type Table struct {
	cols       []table.Column
	rows       []table.Row
	focused    bool
	height     int
	width      int
	clients    map[string]*ty.JsonClient
	ts         table.Styles
	mu         *sync.RWMutex
	logChannel chan ty.Log
}

type MuteTable struct {
	cols        []table.Column
	rows        []table.Row
	height      int
	width       int
	micMute     bool
	speakerMute bool
	ts          table.Styles
	mu          *sync.RWMutex
	logChannel  chan ty.Log
}

func setUpTable(logChan chan ty.Log) (table.Model, *Table) {
	tV := &Table{
		cols: []table.Column{
			{Title: "Name", Width: 5},
			{Title: "Call", Width: 5},
			{Title: "Group", Width: 5},
			{Title: "ClientId", Width: 0},
			{Title: "GroupId", Width: 0},
		},
		rows:       []table.Row{},
		focused:    false,
		height:     5,
		clients:    make(map[string]*ty.JsonClient),
		mu:         &sync.RWMutex{},
		logChannel: logChan,
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

func setUpMuteTable(logChan chan ty.Log) (table.Model, *MuteTable) {
	mTV := &MuteTable{
		cols: []table.Column{
			{Title: "Mic", Width: 7},
			{Title: "Speaker", Width: 7},
		},
		rows:        []table.Row{},
		height:      1,
		micMute:     false,
		speakerMute: false,
		mu:          &sync.RWMutex{},
		logChannel:  logChan,
	}

	mTa := table.New(
		table.WithColumns(mTV.cols),
		table.WithRows(mTV.rows),
		table.WithHeight(mTV.height),
		table.WithFocused(false),
	)

	ts := table.DefaultStyles()
	ts.Header = ts.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderRight(true).
		BorderForeground(purple.GetForeground()).
		Bold(false)
	mTa.SetStyles(ts)
	mTV.ts = ts

	return mTa, mTV
}

func (mT *MuteTable) GetFrameSize() int {
	mT.mu.RLock()
	defer mT.mu.RUnlock()

	return mT.ts.Header.GetHorizontalFrameSize()
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

func (mT *MuteTable) SetMute(toMute string) {
	mT.mu.Lock()
	defer mT.mu.Unlock()

	mT.logChannel <- t.Log{Text: fmt.Sprintf("muteTable: SetMute called with toMute=%s, micMute=%v, speakerMute=%v", toMute, mT.micMute, mT.speakerMute)}
	switch toMute {
	case ty.Microphone:
		if mT.micMute {
			mT.logChannel <- t.Log{Text: "muteTable: Unmuting Microphone"}
			mT.cols[0].Title = noCol.Render(ty.Microphone)
			mT.micMute = false
			mT.logChannel <- t.Log{Text: "muteTable: Microphone is now unmuted"}
			return
		}
		mT.logChannel <- t.Log{Text: "muteTable: Muting Microphone"}
		mT.cols[0].Title = red.Render(ty.Microphone)
		mT.micMute = true
		mT.logChannel <- t.Log{Text: "muteTable: Microphone is now muted"}

	case ty.Speaker:
		if mT.speakerMute {
			mT.logChannel <- t.Log{Text: "muteTable: Unmuting Speaker and Microphone"}
			mT.cols[0].Title = noCol.Render(ty.Microphone)
			mT.cols[1].Title = noCol.Render(ty.Speaker)
			mT.speakerMute, mT.micMute = false, false
			mT.logChannel <- t.Log{Text: "muteTable: Speaker and Microphone are now unmuted"}
			return
		}
		mT.logChannel <- t.Log{Text: "muteTable: Muting Speaker and Microphone"}
		mT.cols[0].Title = red.Render(ty.Microphone)
		mT.cols[1].Title = red.Render(ty.Speaker)
		mT.speakerMute, mT.micMute = true, true
		mT.logChannel <- t.Log{Text: "muteTable: Speaker and Microphone are now muted"}
	}
}

func (t *Table) ConvertClientsToRows(clientToBlink string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rows = t.rows[:0]

	for _, client := range t.clients {
		if client.ClientId == clientToBlink {
			t.rows = append(t.rows, []string{
				green.Render(client.Name),
				green.Render(client.CallState),
				green.Render(client.GroupName),
				client.ClientId,
				client.GroupId,
			})
			continue
		}
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

func (t *Table) GetClient(clientId string) *ty.JsonClient {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.clients[clientId]
}

func (t *Table) GetCallState(clientId string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.clients[clientId].CallState
}

func (t *Table) Empty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.clients) > 0 {
		return false
	}

	return true
}
