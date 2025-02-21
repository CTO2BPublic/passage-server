package eventdriver

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

type Event struct {
	ID            string            `gorm:"primaryKey" json:"id" example:"0d2dab7cdcb4cf1d"` // Pod ID
	CreatedAt     time.Time         `swaggerignore:"true" json:"createdAt,omitempty"`
	UpdatedAt     time.Time         `swaggerignore:"true" json:"updatedAt,omitempty"`
	TransactionID string            `gorm:"index" json:"transactionId" example:"0d2dab7cdcb4cf1d"`
	ParentID      string            `gorm:"index" swaggerignore:"true" json:"parentId"`
	ParentType    EventParentType   `gorm:"index" swaggerignore:"true" json:"parentType"`
	Tenant        string            `gorm:"index" swaggerignore:"true" json:"tenant"`
	Attributes    EventAttributes   `gorm:"embedded;embeddedPrefix:attributes_" json:"attributes"`
	Message       string            `json:"message"`
	Data          datatypes.JSONMap `swaggerignore:"true" json:"data"`
}

type EventAttributes struct {
	Source string    `gorm:"index" example:"sview-hook" json:"source"`
	Type   string    `gorm:"index" example:"gitlab-mr" json:"type"`
	Date   time.Time `gorm:"index" example:"2022-02-28 18:03:49.750647+00" json:"date"`
	Author string    `gorm:"index" example:"john.doe" json:"author"`
}

type EventParentType string

const (
	EventParentApplication EventParentType = "application"
	EventParentParameter   EventParentType = "parameter"
	EventParentSystem      EventParentType = "system"
	EventParentSecurity    EventParentType = "security"
)

func (a *EventAttributes) UnmarshalJSON(data []byte) error {
	// Define a temporary struct to capture the JSON structure
	type TempEventAttributes struct {
		Source string `json:"source"`
		Type   string `json:"type"`
		Date   string `json:"date"`
		Author string `json:"author"`
	}

	// Unmarshal JSON into the temporary struct
	var tempAttr TempEventAttributes
	if err := json.Unmarshal(data, &tempAttr); err != nil {
		return err
	}

	// Try parsing the date as the first format: "2024-02-29T23:12:56.755" (Python events)
	parsedDate, err := time.Parse("2006-01-02T15:04:05.999999999", tempAttr.Date)
	if err != nil {
		// If parsing fails, try parsing as the second format: "2024-03-01T10:58:55.73124427Z" (Go events)
		parsedDate, err = time.Parse(time.RFC3339Nano, tempAttr.Date)
		if err != nil {
			return err
		}
	}

	// Convert parsed date to UTC timezone
	a.Date = parsedDate.UTC()
	a.Source = tempAttr.Source
	a.Type = tempAttr.Type
	a.Author = tempAttr.Author

	return nil
}
