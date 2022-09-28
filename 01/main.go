package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
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
	// -- driver and string for sqlite
	// driver           = "sqlite3"
	// connectionString = "movies.sqlite"

	// -- driver and string for postgres
	driver           = "postgres"
	connectionString = "user=postgres password=postgres dbname=postgres sslmode=disable"
)

func main() {
	db, err := sql.Open(driver, connectionString)
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("select * from movies where rating > 4;")
	if err != nil {
		panic(err)
	}

	var movies []Movie

	for rows.Next() {
		var (
			id      int
			title   string
			genre   string
			revenue string
			rating  int
		)

		if err := rows.Scan(&id, &title, &genre, &revenue, &rating); err != nil {
			panic(err)
		}

		movies = append(movies, Movie{id, title, genre, revenue, rating})
	}

	fmt.Println("🎥 Movies with rating over 4 stars:")
	for _, movie := range movies {
		fmt.Print(movie)
	}
	fmt.Printf("%d movies\n", len(movies))
}
