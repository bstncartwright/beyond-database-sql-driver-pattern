// Package driver defines interfaces to be implemented by an event bus
// drivers as used by package bus.
package driver

import (
	"context"
	"strings"
)

//go:generate mockgen -source=./driver.go -destination=./mocks/mock_driver.go
// TODO: remove this when we migrate to go1.18 in `go.mod`
type any interface{}

// Message reflects the associated topic type-safe struct
type Message any

// Driver is the interface to be implemented by a event bus driver.
type Driver interface {
	// OpenConnector will return a connector where further connections can be made
	OpenConnector() (Connector, error)
}

// Topic type safe struct
type Topic struct {
	Name     string
	Type     interface{}
	Exchange string
	Consumer Consume
}

// Connector is the interface to provide a connection to an event bus.
type Connector interface {
	// Connect opens a connection with the event bus on the given context.
	Connect() (Conn, error)
}

// Consume provides a type of function for consuming messages. The type for msg
// is determined by the driver, and thus the driver's documentation
// should be referenced on what type to assert msg as in order to work with it.
type Consume func(msg Message) error

// Conn is the interface for an open event bus connection.
type Conn interface {
	// Push emits a message onto the event bus in a type safe way (recommended)
	Push(ctx context.Context, topic Topic, message Message) error

	// Subscribe will subscribe to a topics based on the provided consumers map,
	// calling the consume function when a message is received on that topic.
	// There can only be one subscription open per event bus connector.
	Subscribe(ctx context.Context, topics []Topic) error
}

// Resource describes the first delimitation which should be the resource type
func (t Topic) Resource() string {
	s := strings.Split(t.Name, ".")
	return s[0]
}

// Action describes the middle delimitation which should be the verb acting on the resource
func (t Topic) Action() string {
	s := strings.Split(t.Name, ".")
	return s[1]
}
