package driver_nats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	config := getConfig()

	assert.Equal(t, config.NATSSubject, "battlefield")
	assert.Equal(t, config.NATSURL, "tcp://localhost:4222")
}
