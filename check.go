package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgresql://freelance:freelance@localhost:5433/freelance_market?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT from_user_id, to_user_id, content FROM messages")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var from, to, content string
		rows.Scan(&from, &to, &content)
		fmt.Printf("Message: %s -> %s: %s\n", from, to, content)
		count++
	}
	fmt.Printf("Total messages: %d\n", count)
}
