package domain

import (
	"time"
)

type ReportEntityType string

const (
	ReportEntityItem ReportEntityType = "ITEM"
	ReportEntityUser ReportEntityType = "USER"
	ReportEntityChat ReportEntityType = "CHAT"
)

type ReportType string

const (
	ReportTypeSpam          ReportType = "SPAM"
	ReportTypeInappropriate ReportType = "INAPPROPRIATE"
	ReportTypeFraud         ReportType = "FRAUD"
	ReportTypeOther         ReportType = "OTHER"
)

type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "PENDING"
	ReportStatusResolved  ReportStatus = "RESOLVED"
	ReportStatusDismissed ReportStatus = "DISMISSED"
)

type Report struct {
	ID                 string           `json:"id"`
	ReporterID         string           `json:"reporter_id"`
	ReporterName       string           `json:"reporter_name,omitempty"`
	ReportedEntityID   string           `json:"reported_entity_id"`
	ReportedEntityName string           `json:"reported_entity_name,omitempty"`
	EntityType         ReportEntityType `json:"entity_type"`
	Type               ReportType       `json:"type"`
	Description        string           `json:"description"`
	AttachmentURL      string           `json:"attachment_url,omitempty"`
	Status             ReportStatus     `json:"status"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}
