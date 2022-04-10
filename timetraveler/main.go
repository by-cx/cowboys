package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/by-cx/cowboys/common"
	"github.com/by-cx/cowboys/driver_nats"
)

// TimeTraveler records what's happening on the battlefield.
type TimeTraveler struct {
	Driver        common.Driver
	Cowboys       common.Cowboys
	TimeToLeaveCh chan bool
}

// Return list of enemies that are still alive
func (t *TimeTraveler) aliveCowboys() []string {
	alive := []string{}
	for name, cowboy := range t.Cowboys {
		if cowboy.Health > 0 {
			alive = append(alive, name)
		}
	}
	return alive
}

func (t *TimeTraveler) formatCowboy(cowboy common.Cowboy) string {
	return fmt.Sprintf("%s (D%d H%d/%d)", cowboy.Name, cowboy.Damage, cowboy.Health, cowboy.MaxHealth)
}

func (t *TimeTraveler) formatTickMessage(message common.Message) string {
	aliveCowboys := t.aliveCowboys()
	return fmt.Sprintf("new tick, alive cowboys I know about: %s", strings.Join(aliveCowboys, ", "))
}

func (t *TimeTraveler) formatShotMessage(message common.Message) string {
	return fmt.Sprintf("shot from %s (%d), target: %s", message.Source, message.ShotValue, t.formatCowboy(message.Cowboy))
}

func (t *TimeTraveler) formatStatusMessage(message common.Message) string {
	return fmt.Sprintf("status update: %s", t.formatCowboy(message.Cowboy))
}

func (t *TimeTraveler) formatCorruptionMessage(message common.Message) string {
	return "corruption detected, universe is collapsing"
}

// Prints all messages to the stdout
func (t *TimeTraveler) Handler(message common.Message) {
	// In tick message we check who's alive
	if message.Type == common.MessageTypeTick {
		log.Printf("(%s) TICK %d: %s", message.Source, message.Tick, t.formatTickMessage(message))

		// Print status of alive cowboys we know about
		if len(t.Cowboys) > 0 {
			aliveCowboys := t.aliveCowboys()
			if len(aliveCowboys) == 0 {
				log.Printf("(%s) TICK %d: all cowboys are dead", message.Source, message.Tick)

				t.TimeToLeaveCh <- true
			} else if len(aliveCowboys) == 1 {
				log.Printf("(%s) TICK %d: we have the winner! %s", message.Source, message.Tick, aliveCowboys[0])

				t.TimeToLeaveCh <- true
			}
		}
	} else if message.Type == common.MessageTypeShoot {
		log.Printf("(%s) TICK %d: %s", message.Source, message.Tick, t.formatShotMessage(message))
	} else if message.Type == common.MessageTypeStatus {
		log.Printf("(%s) TICK %d: %s", message.Source, message.Tick, t.formatStatusMessage(message))

		// Update local Cowboys storage with latest info
		t.Cowboys[message.Source] = message.Cowboy
	} else if message.Type == common.MessageTypeCorruption {
		log.Printf("(%s) TICK %d: %s", message.Source, message.Tick, t.formatCorruptionMessage(message))
	} else {
		log.Println("unknown message type detected")
	}
}

func main() {
	config := getConfig()
	var err error

	timeTraveler := TimeTraveler{
		TimeToLeaveCh: make(chan bool, 1),
		Cowboys:       make(common.Cowboys),
	}

	// Initiate the message driver
	if config.Driver == "nats" {
		timeTraveler.Driver, err = driver_nats.Init(timeTraveler.Handler)
		if err != nil {
			log.Println("Driver initiation failed:", err)
		}
	} else {
		log.Fatalf("unknown driver %s", config.Driver)
	}

	<-timeTraveler.TimeToLeaveCh
}
