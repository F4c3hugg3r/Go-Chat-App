package chat

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ty "github.com/F4c3hugg3r/Go-Chat-Server/pkg/shared"
)

func TestReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	client := Client{
		Name:      ClientName,
		ClientId:  ClientId,
		clientCh:  make(chan *ty.Response, 100),
		Active:    false,
		authToken: AuthToken,
		lastSign:  time.Now().UTC().Add(-time.Hour),
	}

	client.clientCh <- &DummyResponse

	rsp, err := client.Receive(ctx)
	assert.Nil(t, err)
	assert.Equal(t, rsp, &DummyResponse)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		rsp, err = client.Receive(ctx)
	}()

	cancel()
	wg.Wait()
	assert.Nil(t, rsp)
	assert.ErrorIs(t, err, ty.ErrTimeoutReached)
}

func TestSend(t *testing.T) {
	clientInactive := Client{
		Name:      ClientName,
		ClientId:  ClientId,
		clientCh:  nil,
		Active:    false,
		authToken: AuthToken,
		lastSign:  time.Now().UTC().Add(-time.Hour),
		chClosed:  true,
	}

	err := clientInactive.Send(&DummyResponse)
	assert.ErrorIs(t, err, ty.ErrChannelClosed)

	clientActive := Client{
		Name:      ClientName,
		ClientId:  ClientId,
		clientCh:  make(chan *ty.Response, 100),
		Active:    false,
		authToken: AuthToken,
		lastSign:  time.Now().UTC(),
		chClosed:  false,
	}

	err = clientActive.Send(&DummyResponse)
	assert.Nil(t, err)
}

func TestIsIdle(t *testing.T) {
	clientInactive := Client{
		Name:      ClientName,
		ClientId:  ClientId,
		clientCh:  nil,
		Active:    false,
		authToken: AuthToken,
		lastSign:  time.Now().UTC().Add(-time.Hour),
		chClosed:  true,
	}

	result := clientInactive.Idle(time.Minute * 30)
	assert.True(t, result)

	clientActive := Client{
		Name:      ClientName,
		ClientId:  ClientId,
		clientCh:  make(chan *ty.Response, 100),
		Active:    false,
		authToken: AuthToken,
		lastSign:  time.Now().UTC(),
		chClosed:  false,
	}

	result = clientActive.Idle(time.Minute * 30)
	assert.False(t, result)
}
