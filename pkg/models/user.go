package models

import "fmt"

type UserProfile struct {
	Id       string              `gorm:"primaryKey" json:"id"`
	Username string              `json:"username"`
	Settings UserProfileSettings `json:"settings" gorm:"embedded;embeddedPrefix:settings_"`
}

type UserProfileSettings struct {
	ProviderUsernames
}

type ProviderUsernames struct {
	ProviderUsernames map[string]string `json:"providerUsernames" gorm:"serializer:json"`
}

func (p *UserProfile) Validate() error {
	missingKeys := []string{}
	for username, value := range p.Settings.ProviderUsernames.ProviderUsernames {
		if value == "" {
			missingKeys = append(missingKeys, username)
		}
	}
	if len(missingKeys) > 0 {
		return fmt.Errorf("missing username: %s", missingKeys)
	}
	return nil
}

func (p *UserProfile) GetUser() User {
	return User{
		Id:       p.Id,
		Username: p.Username,
	}
}

type User struct {
	Id       string   `gorm:"primaryKey" json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles,omitempty"`
}
