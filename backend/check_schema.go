package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://dealna_user:amro123@98.92.82.224:5432/dealna_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT column_name, ordinal_position FROM information_schema.columns WHERE table_name = 'profile' ORDER BY ordinal_position")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var pos int
		if err := rows.Scan(&name, &pos); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d: %s\n", pos, name)
	}
}
