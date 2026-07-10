package marketplace

import (
	"github.com/go-chi/chi/v5"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/adapters/primary/giveaway"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
)

// Routes groups Purchase and Transaction routes under /purchases and /transactions.
type Routes struct {
	Purchase    *giveaway.PurchaseHandler
	Transaction *giveaway.TransactionHandler
	logger      middleware.StructuredLoggerInterface
}

func NewRoutes(
	purchaseH *giveaway.PurchaseHandler,
	transactionH *giveaway.TransactionHandler,
	logger middleware.StructuredLoggerInterface,
) *Routes {
	return &Routes{
		Purchase:    purchaseH,
		Transaction: transactionH,
		logger:      logger,
	}
}

func (r *Routes) Register(router chi.Router) {
	// Purchase Requests
	router.Route("/purchases", func(rg chi.Router) {
		rg.Get("/me", r.Purchase.GetMyRequests)
		rg.Get("/items/{itemId}/requests", r.Purchase.ListRequests)

		// Restricted to non-limited students
		rg.Group(func(rg chi.Router) {
			rg.Use(middleware.ForbidRole("LIMITED_STUDENT", r.logger))
			
			// Providers cannot create purchase requests (act as buyers)
			rg.With(middleware.ForbidRole("PROVIDER", r.logger)).Post("/items/{itemId}/request", r.Purchase.CreateRequest)
			
			rg.Post("/items/{itemId}/requests/{requestId}/accept", r.Purchase.AcceptRequest)
			rg.Post("/items/{itemId}/requests/{requestId}/reject", r.Purchase.RejectRequest)
			rg.Post("/items/{itemId}/requests/{requestId}/cancel", r.Purchase.CancelRequest)
		})
	})

	// Transactions
	router.Route("/transactions", func(rg chi.Router) {
		// Restricted to non-limited students
		rg.Group(func(rg chi.Router) {
			rg.Use(middleware.ForbidRole("LIMITED_STUDENT", r.logger))
			rg.Post("/{transactionId}/confirm-seller", r.Transaction.ConfirmSeller)
			
			// Providers cannot act as buyers
			rg.With(middleware.ForbidRole("PROVIDER", r.logger)).Post("/{transactionId}/confirm-buyer", r.Transaction.ConfirmBuyer)
			
			rg.Post("/{transactionId}/cancel", r.Transaction.CancelTransaction)
		})
	})
}

