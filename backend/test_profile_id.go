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
	// Try fetching a profile_id
	var pid string
	err = pool.QueryRow(ctx, "SELECT profile_id FROM profile LIMIT 1").Scan(&pid)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Using profile_id: %s\n", pid)
	prof, err := repo.GetProfileByUserID(ctx, pid) // Pass profile_id to GetProfileByUserID
	if err != nil {
		fmt.Printf("Error fetching with profile_id: %v\n", err)
	} else {
		fmt.Printf("Success? %s\n", prof.DisplayName)
	}
}
