package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/by-cx/cowboys/driver_nats"
	"github.com/by-cx/cowboys/types"
)

// CowboyFight represents inner state of each cowboy and implements the figting logic.
type CowboyFight struct {
	Cowboy  types.Cowboy
	Driver  types.Driver
	Enemies Cowboys
	ExitCh  chan bool

	tick int // Unit of time used to synchronize the whole cluster of cowboys
}

// Loads cowboy specs from the persistent storage based on his name.
// Return representation of myself, my enemies and error if there is any.
func cowboyLoader(name string) (types.Cowboy, Cowboys, error) {
	cowboys := Cowboys{}

	body, err := ioutil.ReadFile("cowboys.js")
	if err != nil {
		return types.Cowboy{}, Cowboys{}, fmt.Errorf("loading cowboys.js error: %v", err)
	}

	err = json.Unmarshal(body, &cowboys)
	if err != nil {
		return types.Cowboy{}, Cowboys{}, fmt.Errorf("parsing cowboys.js error: %v", err)
	}

	enemies := make(Cowboys)
	var myself types.Cowboy

	for _, cowboy := range cowboys {
		if cowboy.Name == name {
			myself = cowboy
		} else {
			enemies[cowboy.Name] = cowboy
		}
	}

	return myself, enemies, fmt.Errorf("cowboy %s not found", name)
}

// Receives a shot from another cowboy
func (c *CowboyFight) receiveShot(message types.Message) {
	if message.Cowboy.Name == c.Cowboy.Name {
		c.Cowboy.Damage -= message.ShotValue
		c.shareStatus()
	}
}

// Return list of enemies that are still alive
func (c *CowboyFight) aliveEnemies() []string {
	aliveEnemies := []string{}
	for name, cowboy := range c.Enemies {
		if cowboy.Health > 0 {
			aliveEnemies = append(aliveEnemies, name)
		}
	}
	return aliveEnemies
}

// Shot another cowboy
func (c *CowboyFight) shoot() {
	aliveEnemies := c.aliveEnemies()

	rand.Seed(time.Now().Unix())
	enemy := aliveEnemies[rand.Intn(len(aliveEnemies))]

	damage := rand.Intn(c.Cowboy.Damage + 1)

	c.Driver.SendMessage(types.Message{
		Source:    c.Cowboy.Name,
		Type:      types.MessageTypeShoot,
		Tick:      c.tick,
		Cowboy:    c.Enemies[enemy],
		ShotValue: damage,
	})
}

func (c *CowboyFight) shareStatus() {
	c.Driver.SendMessage(types.Message{
		Source:    c.Cowboy.Name,
		Type:      types.MessageTypeStatus,
		Tick:      c.tick,
		Cowboy:    c.Cowboy,
		ShotValue: 0,
	})
}

// Processes incoming message from the universe. It's runs main loop of cowboy.
func (c *CowboyFight) handler(message types.Message) {
	// Process ticks
	if message.Type == types.MessageTypeTick {
		c.tick = message.Tick

		if message.Tick > 0 {
			if message.Tick%2 == 0 { // even is time to shoot
				c.shoot()

			} else { // odd is time to check what happend
				// Check if I am the last one
				if len(c.aliveEnemies()) == 0 {
					log.Println("Victory!")
					c.ExitCh <- true
				}

				// Check if I am dead
				if c.Cowboy.Health <= 0 {
					c.ExitCh <- true
				}
			}
		}
	}

	// Store status update of my enemies
	if message.Type == types.MessageTypeStatus {
		if message.Cowboy.Name != c.Cowboy.Name {
			c.Enemies[message.Cowboy.Name] = message.Cowboy
		}
	}

	// Receive a shot for yourself but also for your enemies
	if message.Type == types.MessageTypeShoot {
		c.receiveShot(message)
	}
}

func main() {
	config := getConfig()
	var driver types.Driver
	cowboyFight := CowboyFight{}

	// Load info about our cowboy
	cowboy, enemies, err := cowboyLoader(config.CowboyIdent)
	if err != nil {
		log.Fatal(err)
	}

	// Initiate the message driver
	if config.Driver == "nats" {
		driver, err = driver_nats.Init(cowboyFight.handler)
		if err != nil {
			log.Println("Driver initiation failed:", err)
		}
	} else {
		log.Fatalf("unknown driver %s", config.Driver)
	}

	// Continue with initiating CowboyFight instance
	cowboyFight.Cowboy = cowboy
	cowboyFight.Driver = driver
	cowboyFight.Enemies = enemies
	cowboyFight.ExitCh = make(chan bool)
	cowboyFight.shareStatus()

	// Any error in the driver is fatal so we can exit
	go func() {
		errors := driver.GetErrorsChan()
		err := <-errors
		log.Println(err)
		cowboyFight.ExitCh <- true
	}()

	// Check for the fight to be finished
	log.Println("Waiting for dead or unexpected event ..")
	<-cowboyFight.ExitCh

	// Close anything related to the communication driver
	driver.Close()
}
