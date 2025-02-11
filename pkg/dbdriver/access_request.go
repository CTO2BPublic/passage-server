package dbdriver

import (
	"context"

	"github.com/CTO2BPublic/passage-server/pkg/models"

	"gorm.io/gorm"
)

func (d *Database) InsertAccessRequest(ctx context.Context, data models.AccessRequest) error {
	result := d.Engine.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(&data)
	return result.Error
}

func (d *Database) UpdateAccessRequest(ctx context.Context, data *models.AccessRequest) error {
	result := d.Engine.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Updates(data)
	return result.Error
}

func (d *Database) DeleteAccessRequest(ctx context.Context, data models.AccessRequest) error {
	result := d.Engine.WithContext(ctx).Where("id = ?", data.Id).Unscoped().Delete(models.AccessRequest{})
	return result.Error
}

func (d *Database) SelectAccessRequest(ctx context.Context, data models.AccessRequest) (*models.AccessRequest, error) {
	var result models.AccessRequest
	q := d.Engine.WithContext(ctx).First(&result, models.AccessRequest{Id: data.Id})
	return &result, q.Error
}

func (d *Database) SelectAccessRequests(ctx context.Context) (result []models.AccessRequest, err error) {
	q := d.Engine.WithContext(ctx).Find(&result)

	return result, q.Error
}

func (d *Database) AccessRequestExists(ctx context.Context, data models.AccessRequest) (bool, error) {
	var count int64
	err := d.Engine.WithContext(ctx).Model(&models.AccessRequest{}).Where("id = ?", data.Id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
