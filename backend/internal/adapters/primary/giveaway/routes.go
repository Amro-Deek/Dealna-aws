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
		rg.Post("/queue/{itemId}/join", r.Queue.JoinQueue)
		rg.Post("/queue/{itemId}/leave", r.Queue.LeaveQueue)
		rg.Get("/queue/{itemId}/position/{entryId}", r.Queue.GetQueuePosition)

		// Purchase Requests
		rg.Post("/purchase/{itemId}/request", r.Purchase.CreateRequest)
		rg.Get("/purchase/{itemId}/requests", r.Purchase.ListRequests)

		// Transaction
		rg.Post("/transaction/{transactionId}/confirm-seller", r.Transaction.ConfirmSeller)
		
		// Notifications
		rg.Get("/notifications", r.Notification.ListNotifications)
		rg.Post("/notifications/{notificationId}/read", r.Notification.MarkRead)
	})
}
