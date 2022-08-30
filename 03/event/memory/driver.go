package memory

import (
	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/bus"
	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/driver"
)

type BusDriver struct{}

// OpenConnector opens a connection to the event bus
func (r BusDriver) OpenConnector() (driver.Connector, error) {
	conn := connector{}

	// create a connection to ensure it works
	_, err := conn.Connect()
	if err != nil {
		return nil, err
	}

	return &conn, nil
}

func init() {
	bus.Register("memory", &BusDriver{})
}
