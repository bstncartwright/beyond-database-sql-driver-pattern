package memory

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/driver"
)

// global store for in memory bus driver
var store sync.Map // map[topic.Name]chan driver.Message

type conn struct{}

func (c *conn) Push(ctx context.Context, topic driver.Topic, m driver.Message) error {
	// get channel out of sync.Map. if it doesn't exist, create it and store it
	val, _ := store.LoadOrStore(topic.Name, make(chan driver.Message))

	ch, ok := val.(chan driver.Message)
	if !ok {
		// this really can't happen 👀
		return fmt.Errorf("stored value for topic %q is not of correct type", topic.Name)
	}

	// send the message!
	ch <- m
	return nil
}

func (c *conn) Subscribe(ctx context.Context, topics []driver.Topic) error {
	for _, topic := range topics {
		// get channel out of sync.Map. if it doesn't exist, create it and store it
		val, _ := store.LoadOrStore(topic.Name, make(chan driver.Message))

		ch, ok := val.(chan driver.Message)
		if !ok {
			// this really can't happen 👀
			return fmt.Errorf("stored value for topic %q is not of correct type", topic.Name)
		}

		go func(t driver.Topic, b chan driver.Message) {
			for msg := range b {
				if err := t.Consumer(msg); err != nil {
					log.Printf("Error consuming message on topic %q: %s", t.Name, err)
				}
			}
		}(topic, ch)
	}

	<-ctx.Done()
	return ctx.Err()
}
