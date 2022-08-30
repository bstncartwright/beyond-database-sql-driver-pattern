package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/bus"

	_ "github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/rabbit"
	// _ "github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/memory"
)

type Movie struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Genre   string `json:"genre"`
	Revenue string `json:"revenue"`
	Rating  int    `json:"rating"`
}

func (m Movie) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🍿 Movie %d: %s\n", m.ID, m.Title))
	sb.WriteString(fmt.Sprintf(" 🎬 Genre: %s\n", m.Genre))
	sb.WriteString(fmt.Sprintf(" 💰 Revenue: %s\n", m.Revenue))
	sb.WriteString(fmt.Sprintf(" ⭐ Rating: %d/5\n", m.Rating))
	return sb.String()
}

const (
	driver = "rabbit"
)

func main() {
	rand.Seed(time.Now().Unix())

	fmt.Println("🎬 Movie bus demo starting up 🚀")

	// open up the event bus
	eb, err := bus.Open(driver)
	if err != nil {
		log.Fatal("Open bus in subscription: ", err)
	}

	// Set up a subscriber for all movies to print what movie is released
	// maybe in the future to add it to a database??

	go func() {
		subscribeToMovies(eb)
	}()

	// Let's publish a few movies!
	var movies []Movie
	if err := json.Unmarshal([]byte(moviesJSON), &movies); err != nil {
		log.Fatal("Unmarshal static movies JSON: ", err)
	}

	for {
		movie := movies[rand.Intn(len(movies))]
		fmt.Printf("📤 Pushing movie release %d on bus\n\n", movie.ID)

		if err := eb.Push(bus.MovieRelease, "*", bus.MovieReleaseMessage(movie)); err != nil {
			log.Fatal("Unable to push message on bus: ", err)
		}

		time.Sleep(time.Second * 3)
	}
}

func subscribeToMovies(eb *bus.Bus) {
	consumerFunc := func(mrm bus.MovieReleaseMessage) error {
		movie := Movie(mrm)

		fmt.Println("🆕 New movie released!")
		fmt.Println(movie)
		fmt.Println()
		return nil
	}

	if err := eb.RegisterMovieReleaseConsumer(consumerFunc); err != nil {
		log.Fatal("Register movie release consumer: ", err)
	}

	if err := eb.Subscribe(); err != nil {
		log.Fatal("Subscribe to bus: ", err)
	}
}

var moviesJSON = `[{
  "id": 1,
  "title": "I Hate Christian Laettner",
  "genre": "Documentary",
  "revenue": "$2411410.67",
  "rating": 3
}, {
  "id": 2,
  "title": "Nightmare in Las Cruces, A",
  "genre": "Documentary",
  "revenue": "$95918331.39",
  "rating": 5
}, {
  "id": 3,
  "title": "From Beginning to End (Do Começo ao Fim)",
  "genre": "Drama|Romance",
  "revenue": "$7871625.12",
  "rating": 3
}, {
  "id": 4,
  "title": "Stars Above",
  "genre": "(no genres listed)",
  "revenue": "$55862756.36",
  "rating": 2
}, {
  "id": 5,
  "title": "Snow Dogs",
  "genre": "Adventure|Children|Comedy",
  "revenue": "$27808789.65",
  "rating": 2
}, {
  "id": 6,
  "title": "Odd Couple, The",
  "genre": "Comedy",
  "revenue": "$29005584.48",
  "rating": 3
}, {
  "id": 7,
  "title": "Envy (Kiskanmak)",
  "genre": "Drama",
  "revenue": "$22239402.01",
  "rating": 5
}, {
  "id": 8,
  "title": "Earth Is a Sinful Song, The (Maa on syntinen laulu)",
  "genre": "Drama",
  "revenue": "$99695815.35",
  "rating": 4
}, {
  "id": 9,
  "title": "Fletch",
  "genre": "Comedy|Crime|Mystery",
  "revenue": "$11883960.09",
  "rating": 4
}, {
  "id": 10,
  "title": "Computer Chess",
  "genre": "Comedy",
  "revenue": "$62598209.48",
  "rating": 2
}]`
