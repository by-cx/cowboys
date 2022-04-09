package main

import (
	"fmt"
	"log"

	"github.com/by-cx/cowboys/common"
	"github.com/by-cx/cowboys/driver_nats"
)

type TimeTraveler struct {
	Driver        common.Driver
	Cowboys       common.Cowboys
	TimeToLeaveCh chan bool
}

func (t *TimeTraveler) formatCowboy(cowboy common.Cowboy) {
	return fmt.Sprintf("")
}

func (t *TimeTraveler) formatTickMessage(message common.Message) {
	return fmt.Sprintf("")
}

func (t *TimeTraveler) formatShotMessage(message common.Message) {
	return fmt.Sprintf("")
}

func (t *TimeTraveler) formatStatusMessage(message common.Message) {
	return fmt.Sprintf("")
}

func (t *TimeTraveler) formatCorruptionMessage(message common.Message) {
	return fmt.Sprintf("")
}

func (t *TimeTraveler) Handler(message common.Message) {
	log.Printf("(%s) TICK %d: %s")
}

func main() {
	config := getConfig()
	var err error

	timeTraveler := TimeTraveler{
		TimeToLeaveCh: make(chan bool, 1),
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
