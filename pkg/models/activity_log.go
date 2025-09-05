package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ActivityLog struct {
	ID         string    `gorm:"primaryKey" json:"id" example:"0d2dab7cdcb4cf1d"`
	Date       time.Time `gorm:"index" json:"date"`
	Severity   string    `json:"severity"`
	RaisedBy   string    `json:"raisedBy"`
	ApprovedBy string    `json:"approvedBy"`
	Type       string    `json:"type"`
	Role       string    `json:"role"`
	Message    string    `json:"message"`
	RequestID  string    `json:"requestId"`
	EventID    string    `json:"eventId"`
}

func NewActivityLogFromEvent(e Event) (*ActivityLog, error) {

	resource := e.Data["resource"]
	request, valid := resource.(AccessRequest)
	if !valid {
		return nil, fmt.Errorf("unsupported event")
	}

	log := &ActivityLog{
		ID:         uuid.NewString(),
		Date:       e.Attributes.Date,
		Severity:   "info",
		RaisedBy:   request.Status.RequestedBy,
		ApprovedBy: request.Status.ApprovedBy,
		Type:       e.Attributes.Type,
		Role:       request.RoleRef.Name,
		RequestID:  request.Id,
		EventID:    e.ID,
	}
	parts := strings.Split(e.Attributes.Type, "passage.")
	if len(parts) > 0 {
		log.Message = strings.ReplaceAll(parts[1], ".", " ")
	}

	if strings.Contains(log.Message, "Error") {
		log.Severity = "warning"
	}

	return log, nil

}
