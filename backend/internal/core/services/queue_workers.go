package services

import (
	"context"
	"log"
	"time"
)

func (s *QueueService) StartWorkers(ctx context.Context) {
	// Expiry worker every 60 seconds
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := s.ExpireStaleEntries(context.Background())
				if err != nil {
					log.Printf("Error expiring stale queue entries: %v", err)
				}
			}
		}
	}()

	// Hourly completion check worker
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Implement hour-long limits on completion logic if needed
				log.Println("Running hourly cleanup worker...")
			}
		}
	}()
}
