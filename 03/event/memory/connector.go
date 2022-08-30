package memory

import (
	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/driver"
)

type connector struct{}

func (c connector) Connect() (driver.Conn, error) {
	return &conn{}, nil
}

func (c connector) Driver() driver.Driver {
	return BusDriver{}
}
