package dbdriver

import (
	"context"

	"github.com/CTO2BPublic/passage-server/pkg/models"

	"gorm.io/gorm"
)

func (d *Database) InsertEvent(ctx context.Context, data models.Event) error {
	result := d.Engine.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(&data)
	return result.Error
}

func (d *Database) DeleteEvent(ctx context.Context, data models.Event) error {
	result := d.Engine.WithContext(ctx).Where("id = ?", data.ID).Unscoped().Delete(models.Event{})
	return result.Error
}

func (d *Database) SelectEvent(ctx context.Context, data models.Event) (*models.Event, error) {
	var result models.Event
	q := d.Engine.WithContext(ctx).First(&result, models.Event{ID: data.ID})
	return &result, q.Error
}

func (d *Database) SelectEvents(ctx context.Context) (result []models.Event, err error) {
	q := d.Engine.WithContext(ctx).Omit("data").Find(&result)

	return result, q.Error
}
