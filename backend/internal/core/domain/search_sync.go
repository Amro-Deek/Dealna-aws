package domain

import (
	"time"
)

// QdrantItemPayload represents strictly what the Lambda will upsert as metadata into Qdrant.
// It MUST OMIT Title and Description entirely to save RAM in the vector database.
type QdrantItemPayload struct {
	UniversityID string  `json:"university_id"`
	Category     string  `json:"category"`
	Price        float64 `json:"price"`
	Status       string  `json:"status"`
	Condition    string  `json:"condition"`
	IsGiveaway   bool    `json:"is_giveaway"`
}

// SQSItemEventData is the DTO sent to AWS SQS. It includes Title and Description so the Lambda has the text to embed.
type SQSItemEventData struct {
	ItemID      string            `json:"item_id"`
	Title       string            `json:"title,omitempty"`       // Extracted for Embedding
	Description string            `json:"description,omitempty"` // Extracted for Embedding
	Payload     QdrantItemPayload `json:"payload"`
}

// SearchSyncEvent is the envelope payload sent to AWS SQS so the Python Lambda can update Qdrant.
type SearchSyncEvent struct {
	EventID   string           `json:"event_id"`
	Action    string           `json:"action"` // e.g., "create", "update_status", "delete"
	Data      SQSItemEventData `json:"data"`
	Timestamp time.Time        `json:"timestamp"`
}
