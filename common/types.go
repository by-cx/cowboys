package common

import (
	"encoding/json"
)

type Cowboys map[string]Cowboy

// Cowboy represents a single fighter.
type Cowboy struct {
	Name   string `json:"name"`
	Health int    `json:"health"`
	Damage int    `json:"damage"`
}

// MessageHandler is a function that process all incoming
// messages.
type MessageHandler func(Message)

const MessageTypeTick = "tick"
const MessageTypeShoot = "shoot"
const MessageTypeStatus = "status"
const MessageTypeCorruption = "corruption"

// Message represents single event that happend in the universe
type Message struct {
	Source    string `json:"source"`     // source of the message, either universe or name of the cowboy
	Type      string `json:"type"`       // tick, shoot, status (cowboy's status)
	Tick      int    `json:"tick"`       // when the action happened
	Cowboy    Cowboy `json:"cowboy"`     // name of the cowboy this message is related to
	ShotValue int    `json:"shot_value"` // number of points of the shoot type
}

// Return this stuct as JSON encoded into bytes
func (m *Message) Bytes() ([]byte, error) {
	body, err := json.Marshal(m)
	return body, err
}

// Communication driver interface
type Driver interface {
	Close()
	GetErrorsChan() chan error   // returns channel that is used to pass errors out of the code
	SendMessage(message Message) // sends messages out
}
