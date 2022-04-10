package main

import (
	"log"
	"os"
	"testing"

	"github.com/by-cx/cowboys/common"
	"github.com/by-cx/cowboys/driver_dummy"
	"github.com/stretchr/testify/assert"
)

var driver driver_dummy.DummyDriver
var timeTraveler TimeTraveler
var cowboys common.Cowboys

func TestMain(m *testing.M) {
	var err error

	// Initiate the timetraveler and driver to test
	timeTraveler.TimeToLeaveCh = make(chan bool, 1)

	driver = driver_dummy.Init(timeTraveler.Handler)
	timeTraveler.Driver = driver

	// Test cowboys
	_, cowboys, err = common.CowboyLoader("../cowboys.json", "")
	if err != nil {
		log.Fatal(err)
	}

	// Run the tests
	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestTimeTravelerAliveCowboys(t *testing.T) {
	timeTraveler.Cowboys = cowboys
	aliveCowboys := timeTraveler.aliveCowboys()
	assert.Equal(t, 5, len(aliveCowboys))
}

func TestTimeTravelerHandler(t *testing.T) {
	// Not implemented because the output would require a little bit more logic around log.Printf() and log.Println() functions.
}

func TestTimeTravelerFormatCowboy(t *testing.T) {
	line := timeTraveler.formatCowboy(cowboys["Bill"])
	assert.Equal(t, "Bill (D2 H8/0)", line)
}

func TestTimeTravelerFormatTickMessage(t *testing.T) {
	line := timeTraveler.formatTickMessage(common.Message{
		Type:   common.MessageTypeTick,
		Tick:   123,
		Source: "universe",
	})
	assert.Equal(t, "new tick", line)
}

func TestTimeTravelerFormatShotMessage(t *testing.T) {
	line := timeTraveler.formatShotMessage(common.Message{
		Type:      common.MessageTypeShoot,
		Tick:      123,
		ShotValue: 10,
		Cowboy:    cowboys["Sam"],
		Source:    "Bill",
	})
	assert.Equal(t, "shot from Bill (10), target: Sam (D1 H10/0)", line)
}

func TestTimeTravelerFormatStatusMessage(t *testing.T) {
	line := timeTraveler.formatStatusMessage(common.Message{
		Type:   common.MessageTypeStatus,
		Tick:   123,
		Cowboy: cowboys["Sam"],
		Source: "Sam",
	})
	assert.Equal(t, "status update: Sam (D1 H10/0)", line)
}

func TestTimeTravelerFormatCorruptionMessage(t *testing.T) {
	line := timeTraveler.formatCorruptionMessage(common.Message{
		Type:   common.MessageTypeCorruption,
		Tick:   123,
		Source: "Sam",
	})
	assert.Equal(t, "corruption detected, universe is collapsing", line)
}
