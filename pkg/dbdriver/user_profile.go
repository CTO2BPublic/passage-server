package dbdriver

import (
	"context"

	"github.com/CTO2BPublic/passage-server/pkg/models"

	"gorm.io/gorm"
)

func (d *Database) InsertUserProfile(ctx context.Context, data models.UserProfile) error {
	result := d.Engine.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(&data)
	return result.Error
}

func (d *Database) UpdateUserProfile(ctx context.Context, data models.UserProfile) error {
	result := d.Engine.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Updates(&data)
	return result.Error
}

func (d *Database) SelectUserProfile(ctx context.Context, data models.UserProfile) (models.UserProfile, error) {
	var result models.UserProfile
	q := d.Engine.WithContext(ctx).First(&result, models.UserProfile{Id: data.Id})
	return result, q.Error
}

func (d *Database) SelectUserProfiles(ctx context.Context) ([]models.UserProfile, error) {
	var result []models.UserProfile
	q := d.Engine.WithContext(ctx).Find(&result, models.UserProfile{})
	return result, q.Error
}

func (d *Database) UserProfileExists(ctx context.Context, data models.UserProfile) (bool, error) {
	var count int64
	err := d.Engine.WithContext(ctx).Model(&models.UserProfile{}).Where("id = ?", data.Id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
