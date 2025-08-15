package dbdriver

import (
	"context"

	"github.com/CTO2BPublic/passage-server/pkg/models"

	"gorm.io/gorm"
)

func (d *Database) InsertActivityLog(ctx context.Context, data models.ActivityLog) error {
	result := d.Engine.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(&data)
	return result.Error
}

func (d *Database) DeleteActivityLog(ctx context.Context, data models.ActivityLog) error {
	result := d.Engine.WithContext(ctx).Where("id = ?", data.ID).Unscoped().Delete(models.ActivityLog{})
	return result.Error
}

func (d *Database) SelectActivityLog(ctx context.Context, data models.ActivityLog) (*models.ActivityLog, error) {
	var result models.ActivityLog
	q := d.Engine.WithContext(ctx).First(&result, models.ActivityLog{ID: data.ID})
	return &result, q.Error
}

func (d *Database) SelectActivityLogs(ctx context.Context) (result []models.ActivityLog, err error) {
	q := d.Engine.WithContext(ctx).Omit("data").Find(&result)

	return result, q.Error
}
