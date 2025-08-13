package models

import (
	"fmt"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
)

// Access role
type AccessRole struct {
	Id              string            `gorm:"primaryKey" json:"id,omitempty" example:"3b7af992-5a30-4ce1-821b-cac8194a230b"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Tags            []string          `json:"tags" gorm:"serializer:json"`
	Annotations     map[string]string `json:"annotations" gorm:"serializer:json"`
	Providers       []ProviderConfig  `json:"providers" gorm:"serializer:json"` // Multiple access mappings for the role
	ApprovalRuleRef ApprovalRuleRef   `json:"approvalRuleRef" gorm:"embedded;embeddedPrefix:approvalRuleRef_"`
}

type ProviderConfig struct {
	Name          string            `json:"name"`
	RunAsync      bool              `json:"runAsync"`
	Provider      string            `json:"provider"`
	CredentialRef CredentialRef     `json:"credentialRef" gorm:"embedded;embeddedPrefix:credentialRef_"`
	Parameters    map[string]string `json:"parameters" gorm:"serializer:json"`
}

type CredentialRef struct {
	Name string `json:"name,omitempty"`
}

type ApprovalRuleRef struct {
	Name string `json:"name"`
}

type ApprovalRule struct {
	Name             string   `json:"string"`
	AuthorCanApprove bool     `json:"authorCanApprove"`
	Users            []string `json:"users"`
	Groups           []string `json:"groups"`
}

// HasApprovalPermission checks if a user is allowed to approve based on the approval rule.
func (a *AccessRole) HasAccessRolePermissions(user string, groups []string, rules []ApprovalRule) bool {

	// Find the matching approval rule
	var rule *ApprovalRule
	for _, r := range rules {

		log.Debug().
			Str("refName", a.ApprovalRuleRef.Name).
			Str("ruleName", r.Name).
			Msg("Matching approval rule by ref")

		if r.Name == a.ApprovalRuleRef.Name {

			log.Debug().
				Str("username", user).
				Str("groups", strings.Join(groups, ",")).
				Str("rule", fmt.Sprintf("%+v", r)).
				Msg("Found matching approval rule")

			rule = &r
			break
		}
	}

	// Check if the user is explicitly listed
	if slices.Contains(rule.Users, user) {
		return true
	}

	// If no matching rule is found, deny approval
	if rule == nil {
		return false
	}

	// Check if the user belongs to any approved group
	for _, userGroup := range groups {
		if slices.Contains(rule.Groups, userGroup) {
			return true
		}
	}

	// If no conditions match, deny approval
	return false
}

func (a *AccessRole) GetApprovalRule(rules []ApprovalRule) ApprovalRule {

	for _, r := range rules {
		if r.Name == a.ApprovalRuleRef.Name {
			return r
		}
	}

	return ApprovalRule{}
}
