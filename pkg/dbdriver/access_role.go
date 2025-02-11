package dbdriver

import (
	"context"

	"github.com/CTO2BPublic/passage-server/pkg/models"

	"gorm.io/gorm"
)

func (d *Database) InsertRole(ctx context.Context, data models.AccessRole) error {
	result := d.Engine.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(&data)
	return result.Error
}

func (d *Database) UpdateRole(ctx context.Context, data models.AccessRole) error {
	result := d.Engine.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Updates(&data)
	return result.Error
}

func (d *Database) SelectRole(ctx context.Context, data models.AccessRole) (models.AccessRole, error) {
	var result models.AccessRole
	q := d.Engine.WithContext(ctx).First(&result, models.AccessRole{Id: data.Id})
	return result, q.Error
}

func (d *Database) RoleExists(ctx context.Context, data models.AccessRole) (bool, error) {
	var count int64
	err := d.Engine.WithContext(ctx).Model(&models.AccessRole{}).Where("id = ?", data.Id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
