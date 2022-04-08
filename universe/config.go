package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// Config keeps info about configuration of this daemon
type Config struct {
	Driver string `envconfig:"DRIVER" required:"true" default:"nats"`
}

// GetConfig return configuration created based on environment variables
func getConfig() *Config {
	var config Config

	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	return &config
}
