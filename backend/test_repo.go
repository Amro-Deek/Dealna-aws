package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/secondary/persistence/postgres"
)

func main() {
	connString := "postgres://dealna_user:amro123@98.92.82.224:5432/dealna_db?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := postgres.NewPurchaseRequestRepository(pool)
	
	// Query the first item that has a transaction
	var itemID string
	err = pool.QueryRow(ctx, "SELECT item_id FROM public.transaction LIMIT 1").Scan(&itemID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Testing GetPurchaseRequestsByItem for item:", itemID)
	reqs, err := repo.GetPurchaseRequestsByItem(ctx, itemID)
	if err != nil {
		log.Fatal(err)
	}

	b, _ := json.MarshalIndent(reqs, "", "  ")
	fmt.Println(string(b))
}
