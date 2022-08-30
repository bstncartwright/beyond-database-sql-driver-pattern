package bus

import (
	"encoding/json"

	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/driver"
)

var MovieRelease = driver.Topic{
	Name:     "movie.release.*",
	Type:     MovieReleaseMessage{},
	Exchange: "movie",
}

type MovieReleaseMessage struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Genre   string `json:"genre"`
	Revenue string `json:"revenue"`
	Rating  int    `json:"rating"`
}

func (e *Bus) RegisterMovieReleaseConsumer(
	f func(mrm MovieReleaseMessage) error,
) error {
	topic := CreateMovieReleaseTopic(f)
	return e.RegisterConsumer(topic)
}

// CreateMovieReleaseTopic creates a print job topic
func CreateMovieReleaseTopic(f func(MovieReleaseMessage) error) driver.Topic {
	b := func(msg driver.Message) error {
		pj := MovieReleaseMessage{}
		// msg is an any interface, so go to bytes first
		switch s := msg.(type) {
		case []byte:
			// then we can unmarshal
			err := json.Unmarshal([]byte(s), &pj)
			if err != nil {
				return err
			}
		case MovieReleaseMessage:
			pj = s
		}

		return f(pj)
	}

	topic := MovieRelease
	topic.Consumer = b
	return topic
}
