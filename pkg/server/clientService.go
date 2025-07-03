package server

import (
	"context"
	"fmt"
	"time"
)

func (c *Client) Execute(handler *PluginRegistry, msg *Message) (*Response, error) {
	c.setActive(true)
	defer c.setActive(false)

	defer c.updateLastSign()

	return handler.FindAndExecute(msg)
}

// Receive receives responses from the clientCh
func (c *Client) Receive(ctx context.Context) (*Response, error) {
	c.setActive(true)
	defer c.setActive(false)

	defer c.updateLastSign()

	select {
	case rsp, ok := <-c.clientCh:
		if !ok {
			return nil, fmt.Errorf("%w: your channel was deleted, please register again", ErrChannelClosed)
		}

		return rsp, nil

	case <-ctx.Done():
		return nil, fmt.Errorf("%w: get request timed out", ErrTimeoutReached)
	}
}

// Send sends a response to the clientCh
func (c *Client) Send(rsp *Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.chClosed {
		return fmt.Errorf("%w: your channel was deleted, please register again", ErrChannelClosed)
	}

	select {
	case c.clientCh <- rsp:
		fmt.Printf("\n%s -> %s", rsp.Name, c.Name)
		return nil
	default:
		return fmt.Errorf("%w: response couldn't be sent, try again", ErrTimeoutReached)
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
