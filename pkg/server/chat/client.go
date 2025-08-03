package chat

import (
	"context"
	"fmt"
	"sync"
	"time"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

type PluginHandler interface {
	FindAndExecute(message *ty.Message) (*ty.Response, error)
}

// Client is a communication participant who has a name, unique id and
// channel to receive messages
type Client struct {
	Name      string `json:"name"`
	ClientId  string `json:"clientId"`
	GroupName string `json:"groupName"`
	clientCh  chan *ty.Response
	active    bool
	authToken string
	lastSign  time.Time
	mu        sync.RWMutex
	chClosed  bool
	groupId   string
	// key represents opposing clientId and value the current callState
	rtcs map[string]string
}

func (c *Client) Execute(handler PluginHandler, msg *ty.Message) (*ty.Response, error) {
	c.setActive(true)
	defer c.setActive(false)

	defer c.updateLastSign()

	return handler.FindAndExecute(msg)
}

// Receive receives responses from the clientCh
func (c *Client) Receive(ctx context.Context) (*ty.Response, error) {
	c.setActive(true)
	defer c.setActive(false)

	defer c.updateLastSign()

	select {
	case rsp, ok := <-c.clientCh:
		if !ok {
			return nil, fmt.Errorf("%w: your channel was deleted, please register again", ty.ErrChannelClosed)
		}

		return rsp, nil

	case <-ctx.Done():
		return nil, fmt.Errorf("%w: get request timed out", ty.ErrTimeoutReached)
	}
}

// Send sends a response to the clientCh
func (c *Client) Send(rsp *ty.Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.chClosed {
		return fmt.Errorf("%w: your channel was deleted, please register again", ty.ErrChannelClosed)
	}

	select {
	case c.clientCh <- rsp:
		fmt.Printf("\n%s -> %s", rsp.RspName, c.Name)
		return nil
	default:
		return fmt.Errorf("%w: response couldn't be sent, try again", ty.ErrTimeoutReached)
	}
}

// IsIdle checks if the client is inactive
func (c *Client) Idle(timeLimit time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.active && time.Since(c.lastSign) >= timeLimit {
		c.closeChannelRequireLock()
		return true
	}

	return false
}

func (c *Client) RemoveUnconnectedRTCs() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, callState := range c.rtcs {
		if callState != ty.ConnectedFlag {
			delete(c.rtcs, id)
		}
	}
}

func (c *Client) RemoveRTC(clientId string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.rtcs, clientId)
}

func (c *Client) setActive(active bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.active = active
}

func (c *Client) updateLastSign() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastSign = time.Now().UTC()
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closeChannelRequireLock()
}

func (c *Client) closeChannelRequireLock() {
	if !c.chClosed {
		c.chClosed = true
		close(c.clientCh)
	}
}

func (c *Client) GetAuthToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.authToken
}

func (c *Client) GetGroupId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.groupId
}

func (c *Client) SetGroup(g *Group) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.groupId = g.GroupId
	c.GroupName = g.Name
}

func (c *Client) GetCallState(oppId string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	callState, exists := c.rtcs[oppId]
	if !exists {
		return ""
	}
	return callState
}

func (c *Client) SetCallState(oppId string, callState string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	callState, exists := c.rtcs[oppId]
	if exists {
		return fmt.Errorf("%w: setCallState failed for %s -> %s", ty.ErrNotAvailable, c.ClientId, oppId)
	}

	c.rtcs[oppId] = callState
	return nil
}

func (c *Client) UnsetGroup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.groupId = ""
	c.GroupName = ""
}

func (c *Client) GetName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Name
}

func (c *Client) CheckRTC(oppId string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state, ok := c.rtcs[oppId]
	if !ok {
		return "", fmt.Errorf("%v: no connection with opposing Client %s", oppId, ty.ErrNotAvailable)
	}

	return state, nil
}
