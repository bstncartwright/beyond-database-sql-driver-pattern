// Package bus provides a generic interface around different event bus implementations.
//
// The eventbus package must be used in conjunction with a event bus driver.
// This bus/driver package combination follows the same design pattern as
// Golang's standard library database/sql/driver packages.
//
// You can read more about that pattern here:
// https://go.googlesource.com/go/+/refs/heads/master/src/database/sql/doc.txt
// https://eli.thegreenplace.net/2019/design-patterns-in-gos-databasesql-package/
//
// Currently the bus supports pushing and subscribing by topic.
// Further functionality can be added based on business need.
//
package bus

import (
	"context"
	"fmt"
	"sync"

	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/driver"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]driver.Driver)
)

// ErrNoConsumers is a custom error returned on Subscribe when no consumers
// are registered.
var ErrNoConsumers = fmt.Errorf("no consumers registered")

// Register will register a bus driver 🚌 👨‍💼 to be used.
// This should be called in the init() function of the driver implementation.
func Register(name string, driver driver.Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("bus: register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic(fmt.Sprintf("bus: register called twice for driver %s", name))
	}
	drivers[name] = driver
}

// Bus is the handle representation a number of connections to an event bus.
type Bus struct {
	connector driver.Connector
	Topics    []driver.Topic
	cancelSub context.CancelFunc
}

// Open opens an event bus based on the driver name and driver specific
// data source name, consisting of information needed to connect to the event bus.
func Open(driverName string) (*Bus, error) {
	driversMu.RLock()
	driverInstance, ok := drivers[driverName]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("bus: unknown driver %q (forgotten import?)", driverName)
	}

	connector, err := driverInstance.OpenConnector()
	if err != nil {
		return nil, err
	}

	bus := &Bus{
		connector: connector,
	}

	return bus, nil
}

// Subscribe subscribes all registered topics and calls the provided consume function with the message.
func (e *Bus) Subscribe() error {
	if len(e.Topics) < 1 {
		return fmt.Errorf("unable to subscribe: %w", ErrNoConsumers)
	}

	if e.cancelSub != nil {
		return fmt.Errorf("unable to subscribe: subscription already open")
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.cancelSub = cancel

	conn, err := e.connector.Connect()
	if err != nil {
		return err
	}

	return conn.Subscribe(ctx, e.Topics)
}

// RegisterConsumer register consume method
// This method is not thread safe, DO NOT init twice in your code
func (e *Bus) RegisterConsumer(topic driver.Topic) error {
	e.Topics = append(e.Topics, topic)
	return nil
}

// Push pushes a message to a given topic and partition. Extra args are also passed through
// that can be used by the driver if needed.
func (e *Bus) Push(topic driver.Topic, tenant string, message driver.Message) error {
	return e.PushContext(context.Background(), topic, tenant, message)
}

// PushContext pushes a message to the given topic and partition, in context of the given context.
// The driver should implement a check if the context is done to cancel the push.
func (e *Bus) PushContext(
	ctx context.Context,
	topic driver.Topic,
	tenant string,
	message driver.Message,
) error {
	return e.push(ctx, topic, tenant, message)
}

func (e *Bus) push(
	ctx context.Context,
	topic driver.Topic,
	tenant string,
	message driver.Message,
) error {
	// get connected
	c, err := e.connector.Connect()
	if err != nil {
		return err
	}

	return c.Push(ctx, topic, message)
}
