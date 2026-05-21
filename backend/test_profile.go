package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/persistence/postgres"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://dealna_user:amro123@98.92.82.224:5432/dealna_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := postgres.NewUserRepository(pool)
	// Try fetching some profiles
	profiles, err := pool.Query(ctx, "SELECT user_id FROM profile LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}
	defer profiles.Close()

	for profiles.Next() {
		var uid string
		if err := profiles.Scan(&uid); err != nil {
			log.Fatal(err)
		}
		
		prof, err := repo.GetProfileByUserID(ctx, uid)
		if err != nil {
			fmt.Printf("Error fetching %s: %v\n", uid, err)
		} else {
			fmt.Printf("Success fetching %s: %s\n", uid, prof.DisplayName)
		}
	}
}
