package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	client := Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().UTC().Add(-time.Hour),
	}

	client.clientCh <- &dummyResponse

	rsp, err := client.Receive(ctx)
	assert.Nil(t, err)
	assert.Equal(t, rsp, &dummyResponse)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		rsp, err = client.Receive(ctx)
	}()

	cancel()
	wg.Wait()
	assert.Nil(t, rsp)
	assert.ErrorIs(t, err, ErrTimeoutReached)
}

func TestSend(t *testing.T) {
	clientInactive := Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  nil,
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().UTC().Add(-time.Hour),
		chClosed:  true,
	}

	err := clientInactive.Send(&dummyResponse)
	assert.ErrorIs(t, err, ErrChannelClosed)

	clientActive := Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().UTC(),
		chClosed:  false,
	}

	err = clientActive.Send(&dummyResponse)
	assert.Nil(t, err)
}

func TestIsIdle(t *testing.T) {
	clientInactive := Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  nil,
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().UTC().Add(-time.Hour),
		chClosed:  true,
	}

	result := clientInactive.IsIdle(time.Minute * 30)
	assert.True(t, result)

	clientActive := Client{
		Name:      name,
		ClientId:  clientId,
		clientCh:  make(chan *Response, 100),
		Active:    false,
		authToken: authToken,
		lastSign:  time.Now().UTC(),
		chClosed:  false,
	}

	result = clientActive.IsIdle(time.Minute * 30)
	assert.False(t, result)
}
