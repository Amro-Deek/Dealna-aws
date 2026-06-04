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

	q := `
	SELECT pr.request_id, pr.item_id, pr.buyer_id, t.transaction_id, t.transaction_status
	FROM purchase_request pr
	LEFT JOIN public.transaction t ON t.item_id = pr.item_id AND t.buyer_id = pr.buyer_id
	WHERE pr.status IN ('PENDING_TX', 'ACCEPTED')
	LIMIT 5;
	`
	rows, err := pool.Query(ctx, q)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var reqID, itemID, buyerID string
		var txID, txStatus *string
		err := rows.Scan(&reqID, &itemID, &buyerID, &txID, &txStatus)
		if err != nil {
			log.Println("Scan error:", err)
			continue
		}
		
		fmt.Printf("- Req: %s, Item: %s, Buyer: %s\n", reqID, itemID, buyerID)
		if txID != nil {
			fmt.Printf("  TxID: %s, Status: %s\n", *txID, *txStatus)
		} else {
			fmt.Printf("  TxID: NULL\n")
		}
	}
}
