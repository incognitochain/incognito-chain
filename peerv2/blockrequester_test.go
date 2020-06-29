package peerv2

import (
	"context"
	"testing"
	"time"

	"github.com/incognitochain/incognito-chain/peerv2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestDialWithTimeout makes sure gRPC dialing is blocked and has appropriate timeout
func TestDialWithTimeout(t *testing.T) {
	defer configTime()()

	dialer := &mocks.GRPCDialer{}
	hasTimeout := false
	var err error
	dialer.On("Dial", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, err).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		_, ok := ctx.Deadline()
		hasTimeout = ok
	})
	c := BlockRequester{prtc: dialer, stop: make(chan int, 2)}

	go c.keepConnection()
	time.Sleep(200 * time.Millisecond)
	c.stop <- 1

	assert.True(t, hasTimeout)
}
