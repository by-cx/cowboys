package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	config := getConfig()

	assert.Equal(t, config.Driver, "nats")
}
