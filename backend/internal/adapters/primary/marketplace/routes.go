package marketplace

import (
	"github.com/go-chi/chi/v5"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/giveaway"
)

// Routes groups Purchase and Transaction routes under /purchases and /transactions.
type Routes struct {
	Purchase    *giveaway.PurchaseHandler
	Transaction *giveaway.TransactionHandler
}

func NewRoutes(
	purchaseH *giveaway.PurchaseHandler,
	transactionH *giveaway.TransactionHandler,
) *Routes {
	return &Routes{
		Purchase:    purchaseH,
		Transaction: transactionH,
	}
}

func (r *Routes) Register(router chi.Router) {
	// Purchase Requests
	router.Route("/purchases", func(rg chi.Router) {
		rg.Get("/me", r.Purchase.GetMyRequests)
		rg.Post("/items/{itemId}/request", r.Purchase.CreateRequest)
		rg.Get("/items/{itemId}/requests", r.Purchase.ListRequests)
		rg.Post("/items/{itemId}/requests/{requestId}/accept", r.Purchase.AcceptRequest)
		rg.Post("/items/{itemId}/requests/{requestId}/reject", r.Purchase.RejectRequest)
		rg.Post("/items/{itemId}/requests/{requestId}/cancel", r.Purchase.CancelRequest)
	})

	// Transactions
	router.Route("/transactions", func(rg chi.Router) {
		rg.Post("/{transactionId}/confirm-seller", r.Transaction.ConfirmSeller)
		rg.Post("/{transactionId}/confirm-buyer", r.Transaction.ConfirmBuyer)
	})
}
