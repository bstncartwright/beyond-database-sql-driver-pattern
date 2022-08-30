// Package rabbit provides the implementation of the bus/driver interface
// for the RabbitMQ event-bus-app: https://www.rabbitmq.com/tutorials/tutorial-one-go.html
package rabbit

import (
	"fmt"

	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/bus"
	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/driver"
)

type BusDriver struct{}

// OpenConnector opens a connection to the event bus
func (r BusDriver) OpenConnector() (driver.Connector, error) {
	cfg := NewConfig()
	if cfg.Name == "" {
		return nil, fmt.Errorf("bus name variable is not set, each service needs this set in order to declare a queue")
	}
	conn := connector{cfg: cfg}

	// create a connection to ensure it works
	_, err := conn.Connect()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func init() {
	bus.Register("rabbit", &BusDriver{})
}
