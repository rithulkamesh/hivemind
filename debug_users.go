package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Println("DATABASE_URL required")
		os.Exit(1)
	}

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	rows, _ := conn.Query(context.Background(), "SELECT id, email, username FROM users")
	defer rows.Close()

	for rows.Next() {
		var id, email, username string
		rows.Scan(&id, &email, &username)
		fmt.Printf("User: %s (%s) - %s\n", username, email, id)

		// List keys for this user
		krows, _ := conn.Query(context.Background(), "SELECT id, name, key_prefix FROM api_keys WHERE user_id = $1", id)
		defer krows.Close()
		for krows.Next() {
			var kid, kname, kprefix string
			krows.Scan(&kid, &kname, &kprefix)
			fmt.Printf("  Key: %s (%s) - %s\n", kname, kprefix, kid)
		}
	}
}
