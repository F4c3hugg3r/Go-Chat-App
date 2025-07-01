package server

import (
	"context"
	"fmt"
	"time"
)

// Receive receives responses from the clientCh
func (c *Client) Receive(ctx context.Context) (*Response, error) {
	c.setActive(true)
	defer c.setActive(false)

	defer c.updateLastSign()

	select {
	case rsp, ok := <-c.clientCh:
		if !ok {
			return nil, fmt.Errorf("%w: your channel was deleted, please register again", ChannelClosedError)
		}

		return rsp, nil

	case <-ctx.Done():
		return nil, fmt.Errorf("%w: get request timed out", TimeoutReachedError)
	}
}

// Send sends a response to the clientCh
func (c *Client) Send(rsp *Response) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.chClosed {
		return fmt.Errorf("%w: your channel was deleted, please register again", ChannelClosedError)
	}

	select {
	case c.clientCh <- rsp:
		fmt.Printf("\n%s -> %s", rsp.Name, c.Name)
		return nil
	default:
		return fmt.Errorf("%w: response couldn't be sent, try again", TimeoutReachedError)
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

func (c *Client) closeCh() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chClosed = true
	close(c.clientCh)
}

// IsIdle checks if the client is inactive
func (c *Client) IsIdle(timeLimit time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Active && time.Since(c.lastSign) >= timeLimit {
		return true
	}
	return false
}
