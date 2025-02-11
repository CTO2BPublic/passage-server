package models

import (
	"encoding/json"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
)

// ClaimsMap wraps a map[string]interface{} to represent dynamic claims
type ClaimsMap struct {
	Claims                map[string]interface{} `json:"-"`
	*jwt.RegisteredClaims `swaggerignore:"true"`
}

func NewClaimsMap() ClaimsMap {
	return ClaimsMap{
		Claims: make(map[string]interface{}),
	}
}

func NewUnauthenticatedUserClaims() ClaimsMap {
	return ClaimsMap{
		Claims: map[string]interface{}{
			"username": "Default user",
			"groups":   []string{"default group"},
		},
	}
}

// Valid implements the jwt.Claims interface
func (c ClaimsMap) Valid() error {
	// Validate the standard claims
	if c.RegisteredClaims != nil {
		if err := c.RegisteredClaims.Valid(); err != nil {
			return err
		}
	}

	return nil
}

func (c ClaimsMap) GetString(key string) string {
	if value, exists := c.Claims[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}

// GetStringSlice extracts a string slice from ClaimsMap.
func (c ClaimsMap) GetStringSlice(key string) []string {
	if value, exists := c.Claims[key]; exists {

		// Attempt to convert []interface{} to []string
		if strSlice, ok := value.([]interface{}); ok {
			result := make([]string, len(strSlice))
			for i, v := range strSlice {
				if str, ok := v.(string); ok {
					result[i] = str
				} else {
					log.Warn().Msgf("Non-string value in slice: %v", v)
				}
			}
			return result
		}

		// Handle a single string value
		if singleStr, ok := value.(string); ok {
			return []string{singleStr}
		}
	}

	return []string{}
}

// GetMap extracts a map[string]interface{} field from the claims map
func (c ClaimsMap) GetMap(key string) map[string]interface{} {
	if value, exists := c.Claims[key]; exists {
		if mapValue, ok := value.(map[string]interface{}); ok {
			return mapValue
		}
	}
	return map[string]interface{}{}
}

func (c *ClaimsMap) UnmarshalJSON(data []byte) error {
	// Parse into a temporary map
	temp := make(map[string]interface{})
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Extract and unmarshal standard claims into RegisteredClaims
	registeredClaims := jwt.RegisteredClaims{}
	if err := json.Unmarshal(data, &registeredClaims); err == nil {
		c.RegisteredClaims = &registeredClaims
	}

	// Remove standard JWT claims from the map (so they don't duplicate in Claims)
	standardClaims := []string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti"}
	for _, claim := range standardClaims {
		delete(temp, claim)
	}

	// Assign remaining fields to the Claims map
	c.Claims = temp
	return nil
}

func (c ClaimsMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map to hold all claims
	temp := make(map[string]interface{})

	// Marshal the RegisteredClaims into the map
	if c.RegisteredClaims != nil {
		registeredClaimsData, err := json.Marshal(c.RegisteredClaims)
		if err != nil {
			return nil, err
		}

		// Unmarshal the registered claims back into the map
		registeredClaimsMap := make(map[string]interface{})
		if err := json.Unmarshal(registeredClaimsData, &registeredClaimsMap); err != nil {
			return nil, err
		}

		// Add registered claims to the temporary map
		for key, value := range registeredClaimsMap {
			temp[key] = value
		}
	}

	// Add the dynamic claims from the Claims map
	for key, value := range c.Claims {
		temp[key] = value
	}

	// Marshal the combined map into JSON
	return json.Marshal(temp)
}

func (c ClaimsMap) GetProviderUsernamesFromClaim(claim string) map[string]string {

	mappings := map[string]string{}
	traitsMap := c.GetMap(claim)

	// Loop through the traits and extract the values
	for provider, usernames := range traitsMap {
		if usernamesSlice, ok := usernames.([]interface{}); ok && len(usernamesSlice) > 0 {
			if username, ok := usernamesSlice[0].(string); ok {
				mappings[provider] = username
			}
		}
	}

	return mappings
}
