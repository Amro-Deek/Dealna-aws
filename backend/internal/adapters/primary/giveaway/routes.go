package giveaway

import (
	"github.com/Amro-Deek/Dealna-aws/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	Queue        *QueueHandler
	Notification *NotificationHandler
	logger       middleware.StructuredLoggerInterface
}

func NewRoutes(
	queueH *QueueHandler,
	notificationH *NotificationHandler,
	logger middleware.StructuredLoggerInterface,
) *Routes {
	return &Routes{
		Queue:        queueH,
		Notification: notificationH,
		logger:       logger,
	}
}

func (r *Routes) Register(router chi.Router) {
	router.Route("/giveaway", func(rg chi.Router) {
		// Queue
		rg.Get("/queue/me", r.Queue.GetMyQueues)
		rg.Get("/queue/{itemId}/position/{entryId}", r.Queue.GetQueuePosition)
		rg.Get("/queue/{itemId}/entries", r.Queue.GetQueueEntries)
		
		// Restricted to non-limited students
		rg.Group(func(rg chi.Router) {
			rg.Use(middleware.ForbidRole("LIMITED_STUDENT", r.logger))
			
			// Providers cannot act as buyers / join giveaways
			rg.With(middleware.ForbidRole("PROVIDER", r.logger)).Post("/queue/{itemId}/join", r.Queue.JoinQueue)
			
			rg.Post("/queue/{itemId}/leave", r.Queue.LeaveQueue)
			rg.Post("/queue/{itemId}/entries/{entryId}/accept", r.Queue.AcceptTurn)
			rg.Post("/queue/{itemId}/entries/{entryId}/reject", r.Queue.RejectTurn)
			rg.Post("/queue/{itemId}/entries/{entryId}/handoff", r.Queue.InitiateHandoff)
			rg.Post("/queue/{itemId}/entries/{entryId}/complete", r.Queue.ConfirmHandoff)
		})

		// Notifications
		rg.Get("/notifications/unread-count", r.Notification.GetUnreadCount)
		rg.Get("/notifications", r.Notification.ListNotifications)
		rg.Post("/notifications/{notificationId}/read", r.Notification.MarkRead)
	})
}

