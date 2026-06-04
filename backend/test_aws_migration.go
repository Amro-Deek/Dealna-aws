package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connString := "postgres://dealna_user:amro123@98.92.82.224:5432/dealna_db?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	var count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM public.transaction WHERE seller_confirmed = false").Scan(&count)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	
	fmt.Printf("Success! Count: %d\n", count)
}
