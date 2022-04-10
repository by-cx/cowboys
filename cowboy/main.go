package main

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/by-cx/cowboys/common"
	"github.com/by-cx/cowboys/driver_nats"
)

// CowboyFight represents inner state of each cowboy and implements the figting logic.
type CowboyFight struct {
	Cowboy  common.Cowboy
	Driver  common.Driver
	Enemies common.Cowboys
	ExitCh  chan bool

	tick int // Unit of time used to synchronize the whole cluster of cowboys
}

// Receives a shot from another cowboy
func (c *CowboyFight) receiveShot(message common.Message) {
	if message.Cowboy.Name == c.Cowboy.Name {
		c.Cowboy.Health -= message.ShotValue
		log.Printf("TICK %d: I received a %d shot, my health %d\n", c.tick, message.ShotValue, c.Cowboy.Health)
		c.ShareStatus()
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

	log.Printf("TICK %d: alive enemies: %s\n", c.tick, strings.Join(aliveEnemies, ", "))
	return aliveEnemies
}

// Shot another cowboy
func (c *CowboyFight) shoot() {
	aliveEnemies := c.aliveEnemies()

	rand.Seed(time.Now().Unix())
	enemy := aliveEnemies[rand.Intn(len(aliveEnemies))]

	damage := rand.Intn(c.Cowboy.Damage + 1)

	log.Printf("TICK %d: I shoot on %s with %d damage\n", c.tick, enemy, damage)

	c.Driver.SendMessage(common.Message{
		Source:    c.Cowboy.Name,
		Type:      common.MessageTypeShoot,
		Tick:      c.tick,
		Cowboy:    c.Enemies[enemy],
		ShotValue: damage,
	})
}

func (c *CowboyFight) ShareStatus() {
	if c.Cowboy.MaxHealth == 0 {
		c.Cowboy.MaxHealth = c.Cowboy.Health
	}

	c.Driver.SendMessage(common.Message{
		Source:    c.Cowboy.Name,
		Type:      common.MessageTypeStatus,
		Tick:      c.tick,
		Cowboy:    c.Cowboy,
		ShotValue: 0,
	})
}

// Processes incoming message from the universe. It's runs main loop of cowboy.
func (c *CowboyFight) handler(message common.Message) {
	// Process ticks
	if message.Type == common.MessageTypeTick {
		c.tick = message.Tick

		if message.Tick > 0 {
			if message.Tick%2 == 1 { // even is time to shoot
				c.shoot()

			} else { // odd is time to check what happend
				// Check if I am dead
				if c.Cowboy.Health <= 0 {
					log.Printf("TICK %d: I am DEAD!\n", c.tick)
					c.ExitCh <- true
					return
				}

				// Check if I am the last one
				if len(c.aliveEnemies()) == 0 {
					log.Printf("TICK %d: I won!\n", c.tick)
					c.ExitCh <- true
					return
				}
			}
		} else {
			// Share status when tick is 0 because universe is waiting for us
			c.ShareStatus()
		}
	}

	// Store status update of my enemies
	if message.Type == common.MessageTypeStatus {
		if message.Cowboy.Name != c.Cowboy.Name {
			c.Enemies[message.Cowboy.Name] = message.Cowboy
		}
		return
	}

	// Receive a shot for yourself but also for your enemies
	if message.Type == common.MessageTypeShoot {
		c.receiveShot(message)
	}
}

func main() {
	config := getConfig()
	var driver common.Driver
	cowboyFight := CowboyFight{}

	// Load info about our cowboy
	cowboy, enemies, err := common.CowboyLoader(config.CowboysPath, config.CowboyIdent)
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
	cowboyFight.ShareStatus()

	log.Printf("%s has born\n", cowboyFight.Cowboy.Name)

	// Any error in the driver is fatal so we can exit
	go func() {
		errors := driver.GetErrorsChan()
		err := <-errors
		log.Println(err)
		// TODO: send universe corruption event
		cowboyFight.ExitCh <- true
	}()

	// Check for the fight to be finished
	log.Println("Waiting for dead or unexpected event ..")
	<-cowboyFight.ExitCh

	// Close anything related to the communication driver
	driver.Close()
}
