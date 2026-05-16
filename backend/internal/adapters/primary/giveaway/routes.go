package giveaway

import (
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	Queue        *QueueHandler
	Purchase     *PurchaseHandler
	Transaction  *TransactionHandler
	Notification *NotificationHandler
}

func NewRoutes(
	queueH *QueueHandler,
	purchaseH *PurchaseHandler,
	transactionH *TransactionHandler,
	notificationH *NotificationHandler,
) *Routes {
	return &Routes{
		Queue:        queueH,
		Purchase:     purchaseH,
		Transaction:  transactionH,
		Notification: notificationH,
	}
}

func (r *Routes) Register(router chi.Router) {
	router.Route("/giveaway", func(rg chi.Router) {
		// Queue
		rg.Get("/queue/me", r.Queue.GetMyQueues)
		rg.Post("/queue/{itemId}/join", r.Queue.JoinQueue)
		rg.Post("/queue/{itemId}/leave", r.Queue.LeaveQueue)
		rg.Get("/queue/{itemId}/position/{entryId}", r.Queue.GetQueuePosition)
		rg.Get("/queue/{itemId}/entries", r.Queue.GetQueueEntries)
		rg.Post("/queue/{itemId}/entries/{entryId}/accept", r.Queue.AcceptTurn)
		rg.Post("/queue/{itemId}/entries/{entryId}/reject", r.Queue.RejectTurn)
		rg.Post("/queue/{itemId}/entries/{entryId}/handoff", r.Queue.InitiateHandoff)
		rg.Post("/queue/{itemId}/entries/{entryId}/complete", r.Queue.ConfirmHandoff)

		// Purchase Requests
		rg.Get("/purchases/me", r.Purchase.GetMyRequests)
		rg.Post("/purchases/items/{itemId}/request", r.Purchase.CreateRequest)
		rg.Get("/purchases/items/{itemId}/requests", r.Purchase.ListRequests)
		rg.Post("/purchases/items/{itemId}/requests/{requestId}/accept", r.Purchase.AcceptRequest)
		rg.Post("/purchases/items/{itemId}/requests/{requestId}/reject", r.Purchase.RejectRequest)
		rg.Post("/purchases/items/{itemId}/requests/{requestId}/cancel", r.Purchase.CancelRequest)

		// Transaction
		rg.Post("/transactions/{transactionId}/confirm-seller", r.Transaction.ConfirmSeller)
		rg.Post("/transactions/{transactionId}/confirm-buyer", r.Transaction.ConfirmBuyer)
		
		// Notifications
		rg.Get("/notifications", r.Notification.ListNotifications)
		rg.Post("/notifications/{notificationId}/read", r.Notification.MarkRead)
	})
}
