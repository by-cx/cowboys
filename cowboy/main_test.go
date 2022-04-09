package main

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/by-cx/cowboys/common"
	"github.com/by-cx/cowboys/driver_dummy"
	"github.com/stretchr/testify/assert"
)

var driver driver_dummy.DummyDriver
var cowboyFight CowboyFight
var backupCowboy common.Cowboy

const testCowboyName = "John"

func TestMain(m *testing.M) {
	cowboyFight = CowboyFight{}

	driver = driver_dummy.Init(cowboyFight.handler)

	cowboy, enemies, err := common.CowboyLoader("../cowboys.js", testCowboyName)
	if err != nil {
		log.Fatal(err)
	}

	backupCowboy = cowboy
	cowboyFight.Cowboy = cowboy
	cowboyFight.Driver = driver
	cowboyFight.Enemies = enemies
	cowboyFight.ExitCh = make(chan bool)

	go func() {
		err := <-cowboyFight.Driver.GetErrorsChan()
		log.Println("error occurred:", err)
		os.Exit(1)
	}()

	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestCowboyFight(t *testing.T) {
	go cowboyFight.ShareStatus()

	message := <-driver.OutgoingMessageCh

	assert.Equal(t, message.Cowboy.Name, testCowboyName)
	assert.Equal(t, message.Cowboy.Health, 10)
	assert.Equal(t, message.Cowboy.Damage, 1)
}

func TestCowboyFightHandlerZeroTick(t *testing.T) {
	// Test odd tick
	func() {
		driver.InjectMessageCh <- common.Message{
			Source: "universe",
			Type:   common.MessageTypeTick,
			Tick:   0,
		}
	}()

	message := <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeStatus)
	assert.Equal(t, message.Tick, 0)
	assert.Equal(t, message.Cowboy.Name, testCowboyName)
}

func TestCowboyFightHandlerOddTick(t *testing.T) {
	// Test odd tick
	func() {
		driver.InjectMessageCh <- common.Message{
			Source: "universe",
			Type:   common.MessageTypeTick,
			Tick:   1,
		}
	}()

	message := <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeShoot)
	assert.Equal(t, message.Tick, 1)
	assert.NotEqual(t, message.Cowboy.Name, testCowboyName)
}

func TestCowboyFightHandlerEnemiesStatusUpdate(t *testing.T) {
	assert.Equal(t, cowboyFight.Enemies["Bill"].Health, 8)

	func() {
		driver.InjectMessageCh <- common.Message{
			Source: "Bill",
			Type:   common.MessageTypeStatus,
			Tick:   1,
			Cowboy: common.Cowboy{
				Name:   "Bill",
				Health: 1,
				Damage: 2,
			},
		}
	}()

	// TODO: this is opportunistic testing and it deserve a better way
	// The problem is that we don't have any way how to detect that message handler finished his job.
	time.Sleep(1 * time.Second)
	assert.Equal(t, cowboyFight.Enemies["Bill"].Health, 1)
}

func TestCowboyFightHandlerEvenTick(t *testing.T) {
	// Test even tick
	backupEnemies := cowboyFight.Enemies
	cowboyFight.Enemies = make(common.Cowboys)

	func() {
		driver.InjectMessageCh <- common.Message{
			Source: "universe",
			Type:   common.MessageTypeTick,
			Tick:   2,
		}
	}()

	//! This can potentially freeze the testing
	assert.Equal(t, true, <-cowboyFight.ExitCh)

	cowboyFight.Enemies = backupEnemies
}

func TestCowboyFightHandlerShootToDead(t *testing.T) {
	assert.Equal(t, 10, cowboyFight.Cowboy.Health)

	// First shot
	func() {
		driver.InjectMessageCh <- common.Message{
			Source:    "Sam",
			Type:      common.MessageTypeShoot,
			Tick:      1,
			Cowboy:    cowboyFight.Cowboy,
			ShotValue: 6,
		}
	}()

	message := <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeStatus)
	assert.Equal(t, 4, message.Cowboy.Health)
	assert.Equal(t, 4, cowboyFight.Cowboy.Health)

	// Second shot, deadly
	func() {
		driver.InjectMessageCh <- common.Message{
			Source:    "Sam",
			Type:      common.MessageTypeShoot,
			Tick:      1,
			Cowboy:    cowboyFight.Cowboy,
			ShotValue: 6,
		}
	}()

	message = <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeStatus)
	assert.Equal(t, -2, message.Cowboy.Health)
	assert.Equal(t, -2, cowboyFight.Cowboy.Health)

	// Even tick to check if cowboy dies
	func() {
		driver.InjectMessageCh <- common.Message{
			Source: "universe",
			Type:   common.MessageTypeTick,
			Tick:   2,
		}
	}()

	time.Sleep(1 * time.Second)
	exit := false
	// This prevents the channel to freeze the testing
	select {
	case value := <-cowboyFight.ExitCh:
		exit = value
	default:
		exit = false
	}
	assert.Equal(t, true, exit)
}

func TestCowboyFightReceiveShot(t *testing.T) {
	cowboyFight.Cowboy = backupCowboy

	go func() {
		cowboyFight.receiveShot(common.Message{
			Source:    "Sam",
			Type:      common.MessageTypeShoot,
			Tick:      1,
			Cowboy:    cowboyFight.Cowboy,
			ShotValue: 6,
		})
	}()

	time.Sleep(1 * time.Second)
	assert.Equal(t, 4, cowboyFight.Cowboy.Health)
	message := <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeStatus)
}

func TestCowboyFightAliveEnemies(t *testing.T) {
	enemies := cowboyFight.aliveEnemies()
	assert.Equal(t, 4, len(enemies))
	assert.Contains(t, enemies, "Bill")
}

func TestCowboyFightShoot(t *testing.T) {
	go func() {
		cowboyFight.shoot()
	}()

	time.Sleep(time.Second)
	message := <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeShoot)
	assert.GreaterOrEqual(t, message.ShotValue, 0)
}

func TestCowboyFightShareStatus(t *testing.T) {
	cowboyFight.Cowboy = backupCowboy

	go func() {
		cowboyFight.ShareStatus()
	}()

	time.Sleep(time.Second)
	message := <-driver.OutgoingMessageCh
	assert.Equal(t, message.Type, common.MessageTypeStatus)
	assert.Equal(t, message.Cowboy, backupCowboy)
}
