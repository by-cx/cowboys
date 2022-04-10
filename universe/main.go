package main

import (
	"log"
	"time"

	"github.com/by-cx/cowboys/common"
	"github.com/by-cx/cowboys/driver_nats"
)

// Universe gives cowboys space time where they can create their own history.
type Universe struct {
	Cowboys           common.Cowboys
	ExpectedCowboys   int
	SleepBetweenTicks int
	Driver            common.Driver

	tick int
}

// Message handler
func (u *Universe) Handler(message common.Message) {
	if message.Type == common.MessageTypeStatus {
		u.Cowboys[message.Cowboy.Name] = message.Cowboy
	}
}

// Special zero tick
func (u *Universe) doZeroTick() {
	u.Driver.SendMessage(common.Message{
		Source: "universe",
		Type:   common.MessageTypeTick,
		Tick:   0,
	})
}

// Tick once
func (u *Universe) doTick() {
	u.Driver.SendMessage(common.Message{
		Source: "universe",
		Type:   common.MessageTypeTick,
		Tick:   u.tick,
	})

	u.tick += 1
}

// Return number of alive cowboys
func (u *Universe) aliveCowboys() int {
	alive := 0
	for _, cowboy := range u.Cowboys {
		if cowboy.Health > 0 {
			alive += 1
		}
	}

	log.Printf("TICK %d: alive cowboys: %d\n", u.tick, alive)
	return alive
}

// Implementation of time in our universe
func (u *Universe) Ticking() {
	// Let's wait for all cowboys to be ready
	// This is actually a shortcut, I would rather implement this distributed in the cowboy's code but time is ticking
	if !u.ready() {
		log.Fatalln("Something unexpected has happened.")
	}

	// When they are ready we send regular ticks to the battlefield
	for {
		u.doTick()
		time.Sleep(time.Duration(u.SleepBetweenTicks) * time.Second)

		// When only one or none cowboy is alive there is no need for time itself
		if u.aliveCowboys() <= 1 {
			log.Printf("TICK %d: the fight is over, let's do last tick\n", u.tick)
			// The last tick is needed for the cowboys to figure out they won and let them end themselves. Otherwise they are waiting til end of the real universe.
			u.doTick()
			time.Sleep(time.Second) // we give doTick time to send the message otherwise the process ends before it's sent
			return
		}
	}
}

// Returns true when all cowboys are ready
func (u *Universe) ready() bool {
	for {
		log.Printf("TICK %d: waiting for cowboys to be ready (%d/%d)\n", u.tick, len(u.Cowboys), u.ExpectedCowboys)

		if u.ExpectedCowboys == len(u.Cowboys) {
			return true
		}

		u.doZeroTick()
		time.Sleep(time.Second * 1)
	}
}

func main() {
	config := getConfig()
	var err error

	// Load cowboys
	_, cowboys, err := common.CowboyLoader(config.CowboysPath, "")
	if err != nil {
		log.Fatalln(err)
	}

	// Initiate Universe
	universe := Universe{
		Cowboys:           make(common.Cowboys),
		SleepBetweenTicks: config.SleepBetweenTicks,
		ExpectedCowboys:   len(cowboys),
	}

	// Initiate the message driver
	if config.Driver == "nats" {
		universe.Driver, err = driver_nats.Init(universe.Handler)
		if err != nil {
			log.Println("Driver initiation failed:", err)
		}
	} else {
		log.Fatalf("unknown driver %s", config.Driver)
	}

	// Ticking
	// This is blocking call, when universe decides to crash, implode into itself or simply die it will continue
	universe.Ticking()

	log.Println("This universe is not needed anymore")

	// Close anything related to the communication driver
	universe.Driver.Close()
}
