package models

type Credential struct {
	Name       string               `json:"name"`
	FromSecret CredentialFromSecret `json:"fromSecret"`
	Data       map[string]string    `json:"data" gorm:"serializer:json"`
}

type CredentialFromSecret struct {
	Name string `json:"name"`
}

func (c *Credential) GetString(key string) string {
	if value, ok := c.Data[key]; ok {
		return value
	}
	return ""
}
