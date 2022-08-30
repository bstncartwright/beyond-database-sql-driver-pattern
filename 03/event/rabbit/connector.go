package rabbit

import (
	"fmt"

	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/driver"
	amqp "github.com/rabbitmq/amqp091-go"
)

type connector struct {
	cfg config
}

func (c connector) Connect() (driver.Conn, error) {
	str := fmt.Sprintf(
		"%s://%s:%s@%s:%s",
		c.cfg.Scheme,
		c.cfg.Username,
		c.cfg.Password,
		c.cfg.Host,
		c.cfg.Port,
	)
	con, err := amqp.DialConfig(
		str,
		amqp.Config{Properties: amqp.Table{"connection_name": c.cfg.Name}},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to rabbitmq: %w", err)
	}

	ch, err := con.Channel()
	if err != nil {
		return nil, fmt.Errorf("unable to open unique channel: %w", err)
	}

	ret := &rabbit{
		cfg:  c.cfg,
		conn: con,
		ch:   ch,
	}
	return ret, nil
}

func (c connector) Driver() driver.Driver {
	return BusDriver{}
}
