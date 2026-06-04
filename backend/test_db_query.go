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

	// Query an item that has purchase requests with status PENDING_TX or ACCEPTED
	q := `
	SELECT 
	  pr.request_id,
	  pr.status,
	  (
		  SELECT t.transaction_id 
		  FROM public.transaction t 
		  WHERE t.item_id = pr.item_id 
			AND t.buyer_id = pr.buyer_id 
			AND t.transaction_status != 'CANCELLED' 
		  LIMIT 1
	  ) AS transaction_id
	FROM purchase_request pr
	WHERE pr.status IN ('PENDING_TX', 'ACCEPTED')
	LIMIT 5;
	`
	rows, err := pool.Query(ctx, q)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Purchase Requests with PENDING_TX / ACCEPTED:")
	for rows.Next() {
		var reqID string
		var status string
		var txID *string
		err := rows.Scan(&reqID, &status, &txID)
		if err != nil {
			log.Println("Scan error:", err)
			continue
		}
		if txID != nil {
			fmt.Printf("- Request: %s, Status: %s, TxID: %s\n", reqID, status, *txID)
		} else {
			fmt.Printf("- Request: %s, Status: %s, TxID: NULL\n", reqID, status)
		}
	}
}
