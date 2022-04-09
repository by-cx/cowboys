package driver_dummy

import (
	"github.com/by-cx/cowboys/common"
)

// DummyDriver has same interface as ordinary driver but it works without any external depency and
// though can be used effectively for testing.
type DummyDriver struct {
	MessageHandler common.MessageHandler
	errorsCh       chan error

	// Those two attributes are used for testing
	OutgoingMessageCh chan common.Message // This is place where outgoing messages can be obtained
	InjectMessageCh   chan common.Message // This can be used to inject a message into the driver
}

func Init(messageHandler common.MessageHandler) DummyDriver {
	driver := DummyDriver{}
	driver.errorsCh = make(chan error)
	driver.OutgoingMessageCh = make(chan common.Message)
	driver.InjectMessageCh = make(chan common.Message)
	driver.MessageHandler = messageHandler

	go func() {
		for message := range driver.InjectMessageCh {
			driver.MessageHandler(message)
		}
	}()

	return driver
}

func (d DummyDriver) Close() {

}

// GetErrorsChan returns channel that is used to pass errors out of the code
func (d DummyDriver) GetErrorsChan() chan error {
	return d.errorsCh
}

func (d DummyDriver) SendMessage(message common.Message) {
	d.OutgoingMessageCh <- message
}
