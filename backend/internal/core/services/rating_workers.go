package services

import (
	"context"
	"log"
	"time"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/domain"
)

func (s *RatingService) StartRatingReminderWorker(ctx context.Context) {
	// Wait 5 minutes before starting to allow the server to fully initialize
	select {
	case <-ctx.Done():
		return
	case <-time.After(5 * time.Minute):
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	s.runReminderJob(ctx) // Run once on startup

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Rating Reminder Worker...")
			return
		case <-ticker.C:
			s.runReminderJob(ctx)
		}
	}
}

func (s *RatingService) runReminderJob(ctx context.Context) {
	log.Println("Running Rating Reminder Worker...")
	
	// Get transactions completed exactly 3 days ago where the buyer hasn't rated yet
	pendingRatings, err := s.ratingRepo.GetTransactionsToRemind(ctx, 3)
	if err != nil {
		log.Printf("Error getting transactions to remind: %v", err)
		return
	}

	if len(pendingRatings) == 0 {
		return
	}

	log.Printf("Found %d pending ratings to remind buyers about.", len(pendingRatings))

	for _, pending := range pendingRatings {
		itemIDStr := pending.ItemID.String()
		buyerIDStr := pending.BuyerID.String()

		notifCtx := NotificationContext{
			ItemID:       &itemIDStr,
		}

		err := s.notifs.CreateNotification(ctx, buyerIDStr, domain.NotifTypeRatingReminder, notifCtx)
		if err != nil {
			log.Printf("Failed to send rating reminder to user %s: %v", buyerIDStr, err)
		}
	}
}
