package driver_nats

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/by-cx/cowboys/common"
	"github.com/nats-io/nats.go"
)

type NATSDriver struct {
	MessageHandler common.MessageHandler
	errorsCh       chan error

	nc *nats.Conn
}

// GetErrorsChan returns channel that is used to pass errors out of the code
func (n NATSDriver) GetErrorsChan() chan error {
	return n.errorsCh
}

// Close active connection to the NATS server
// We don't care about the error here, we just need to know about it
func (n NATSDriver) Close() {
	err := n.nc.Drain()
	if err != nil {
		log.Println("close connection", err)
	}
}

// SendMessage shares a message with the universe
func (n NATSDriver) SendMessage(message common.Message) {
	config := getConfig()

	body, err := message.Bytes()
	if err != nil {
		n.errorsCh <- fmt.Errorf("publishing message error: %v", err)
		return
	}

	err = n.nc.Publish(config.NATSSubject, body)
	if err != nil {
		n.errorsCh <- fmt.Errorf("publishing message error: %v", err)
		return
	}
}

func (n *NATSDriver) handler(m *nats.Msg) {
	var message common.Message
	err := json.Unmarshal(m.Data, &message)
	if err != nil {
		n.errorsCh <- fmt.Errorf("message handling error: %v", err)
		return
	}

	n.MessageHandler(message)
}

// Init connect to the NATS server and subscribes to configured subject.
func Init(messageHandler common.MessageHandler) (common.Driver, error) {
	config := getConfig()

	driver := NATSDriver{
		MessageHandler: messageHandler,
	}

	var err error
	driver.errorsCh = make(chan error)

	// Connect to the NATS server
	for {
		driver.nc, err = nats.Connect(config.NATSURL)
		if err != nil {
			log.Printf("Can't connect to the NATS server, waiting for 5 seconds before I try it again. (%v)\n", err)
			time.Sleep(time.Second * 5)
			continue
		}

		break
	}

	_, err = driver.nc.Subscribe(config.NATSSubject, driver.handler)
	if err != nil {
		return driver, fmt.Errorf("subscribe to subject error: %v", err)
	}

	return driver, nil
}
