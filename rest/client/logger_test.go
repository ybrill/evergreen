package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerClose(t *testing.T) {
	assert := assert.New(t)
	comm := NewCommunicator("www.foo.com")
	logger, err := comm.GetLoggerProducer(context.Background(), TaskData{}, nil)
	assert.NoError(err)
	assert.NotNil(logger)
	assert.NoError(logger.Close())
	assert.NotPanics(func() {
		assert.NoError(logger.Close())
	})
}
