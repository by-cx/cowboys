package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	os.Setenv("COWBOY_IDENT", "testcowboy")

	config := getConfig()

	assert.Equal(t, config.CowboyIdent, "testcowboy")
	assert.Equal(t, config.Driver, "nats")
}
