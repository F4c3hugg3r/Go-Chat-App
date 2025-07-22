package chat

import (
	"context"
	"fmt"
	"sync"
	"time"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

// Client is a communication participant who has a name, unique id and
// channel to receive messages
type Client struct {
	Name      string
	ClientId  string
	clientCh  chan *ty.Response
	Active    bool
	authToken string
	lastSign  time.Time
	mu        sync.RWMutex
	chClosed  bool
	groupId   string
}

func (c *Client) Execute(handler *PluginRegistry, msg *ty.Message) (*ty.Response, error) {
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

func (c *Client) setActive(active bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Active = active
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

// IsIdle checks if the client is inactive
func (c *Client) Idle(timeLimit time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Active && time.Since(c.lastSign) >= timeLimit {
		c.closeChannelRequireLock()
		return true
	}

	return false
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

func (c *Client) SetGroup(groupId string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.groupId = groupId
}

func (c *Client) UnsetGroup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.groupId = ""
}
