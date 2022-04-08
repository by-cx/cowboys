package driver_nats

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// Config keeps info about configuration of this daemon
type Config struct {
	NATSURL     string `envconfig:"NATS_URL" required:"true" default:"tcp://localhost:4222"`
	NATSSubject string `envconfig:"NATS_SUBJECT" required:"true" default:"battlefield"`
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
