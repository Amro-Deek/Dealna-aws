package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

func main() {
	txID := "12345-67890"
	req := domain.PurchaseRequest{
		RequestID:     "req-1",
		ItemID:        "item-1",
		BuyerID:       "buyer-1",
		Status:        domain.PurchaseRequestAccepted,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		BuyerName:     "Amro",
		BuyerPic:      "pic.jpg",
		TransactionID: &txID,
	}

	b, _ := json.MarshalIndent(req, "", "  ")
	fmt.Println(string(b))
}
