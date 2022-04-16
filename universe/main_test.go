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
var universe Universe
var cowboys common.Cowboys

func TestMain(m *testing.M) {
	var err error

	// Load cowboys from the test file
	_, cowboys, err = common.CowboyLoader("../cowboys.json", "")
	if err != nil {
		log.Fatalln(err)
	}

	// Initiate the universe and driver to test
	universe = Universe{
		cowboys:           make(common.Cowboys),
		SleepBetweenTicks: 1,
		ExpectedCowboys:   len(cowboys),
	}

	driver = driver_dummy.Init(universe.Handler)
	universe.Driver = driver

	// Run the tests
	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestHandler(t *testing.T) {
	universe.Handler(common.Message{
		Source: "BIll",
		Type:   common.MessageTypeStatus,
		Tick:   0,
		Cowboy: cowboys["Bill"],
	})

	assert.Equal(t, cowboys["Bill"].Health, universe.readCowboy("Bill").Health)
}

func TestDoZeroTick(t *testing.T) {
	// Test odd tick
	go func() {
		universe.doZeroTick()
	}()

	message := <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeTick)
	assert.Equal(t, message.Tick, 0)
}

func TestDoTick(t *testing.T) {
	syncCh := make(chan bool, 2)
	go func() {
		universe.doTick()
		syncCh <- true
	}()

	message := <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeTick)
	assert.Equal(t, message.Tick, 0)

	firstTick := message.Tick

	go func() {
		<-syncCh
		universe.doTick()
		syncCh <- true
	}()

	message = <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeTick)
	assert.Equal(t, firstTick+1, message.Tick)

	<-syncCh
}

func TestTicking(t *testing.T) {
	universe.writeCowboys(cowboys)
	ch := make(chan bool, 1)

	go func() {
		universe.Ticking()
		ch <- true
	}()

	// We expect first tick here
	message := <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeTick)
	assert.Equal(t, 2, message.Tick)

	// After all cowboys are dead, we expect one more tick and then stop
	for name := range cowboys {
		cowboy := universe.readCowboy(name)
		cowboy.Health = 0
		universe.writeCowboy(name, cowboy)
	}

	// Check if final tick has been sent
	message = <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeTick)
	assert.Equal(t, 3, message.Tick)

	// There are no messages here, after the last message the code waits for one second and then it finishes because everyone is dead
	assert.True(t, <-ch)
}

func TestReady(t *testing.T) {
	universe.cowboys = make(common.Cowboys)
	ch := make(chan bool, 1)

	go func() {
		ch <- universe.ready()
	}()

	message := <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeTick)
	assert.Equal(t, 0, message.Tick)

	universe.writeCowboys(cowboys)

	assert.True(t, <-ch)
}

func TestUniverseReadCowboy(t *testing.T) {
	newCowboys := cowboys
	newCowboys["John"] = common.Cowboy{
		Name: "Someoneelse",
	}
	universe.writeCowboys(newCowboys)

	assert.Equal(t, "Someoneelse", universe.cowboys["John"].Name)
}

func TestUniverseReadCowboys(t *testing.T) {
	newCowboys := cowboys
	newCowboys["John"] = common.Cowboy{
		Name: "Someoneelse",
	}
	universe.writeCowboys(newCowboys)

	assert.Equal(t, "Someoneelse", universe.readCowboy("John").Name)
}

func TestUniverseWriteCowboy(t *testing.T) {
	newCowboys := cowboys
	newCowboys["John"] = common.Cowboy{
		Name: "Someoneelse",
	}
	universe.writeCowboys(newCowboys)
	universe.writeCowboy("Newone", common.Cowboy{Name: "Anotherone"})
	assert.Equal(t, "Anotherone", universe.readCowboy("Newone").Name)
}

func TestUniverseWriteCowboys(t *testing.T) {
	newCowboys := cowboys
	newCowboys["John"] = common.Cowboy{
		Name: "Someoneelse",
	}
	universe.writeCowboys(newCowboys)
	assert.Equal(t, "Anotherone", universe.readCowboy("Newone").Name)
}
