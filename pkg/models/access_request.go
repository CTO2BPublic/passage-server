package models

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// Access request status constants
const (
	AccessRequestPending  = "Pending"
	AccessRequestApproved = "Approved"
	AccessRequestDenied   = "Denied"
	AccessRequestExpired  = "Expired"
	ProviderStatusGranted = "Granted"
	ProviderStatusRevoked = "Revoked"
	ProviderStatusError   = "Error"
)

// Access request
type AccessRequest struct {
	Id        string               `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time            `gorm:"index" swaggerignore:"true" json:"createdAt"`
	UpdatedAt time.Time            `swaggerignore:"true" json:"updatedAt"`
	DeletedAt *time.Time           `gorm:"index" swaggerignore:"true" json:"deletedAt,omitempty"`
	RoleRef   AccessRoleRef        `gorm:"embedded;embeddedPrefix:roleRef_" json:"roleRef"`
	Details   AccessRequestDetails `gorm:"embedded;embeddedPrefix:details_" json:"details"`
	Status    AccessRequestStatus  `swaggerignore:"true" gorm:"embedded;embeddedPrefix:status_" json:"status"`
}

type AccessRoleRef struct {
	Name string `json:"name" example:"SRE-PU-ACCESS"`
}

type AccessRequestDetails struct {
	Justification string                 `json:"justification" example:"Need to access k8s namespace"`
	Attributes    map[string]interface{} `json:"attributes" gorm:"serializer:json"`
	TTL           string                 `json:"ttl" example:"72h"`
}

type AccessRequestStatus struct {
	Status            string                    `json:"status"`
	ApprovedBy        string                    `json:"approvedBy"`
	RequestedBy       string                    `json:"requestedBy"`
	ApprovalRule      ApprovalRule              `json:"approvalRule" gorm:"serializer:json"`
	ProviderUsernames map[string]string         `json:"providerUsernames" gorm:"serializer:json"`
	ProviderStatuses  map[string]ProviderStatus `json:"providerStatuses" gorm:"serializer:json"`
	ExpiresAt         *time.Time
	Trace             string `json:"trace"`
}

type ProviderStatus struct {
	Action  string `json:"action" example:"Granted"`
	Details string `json:"details" example:"Group: sre-pu-sers"`
	Error   string `json:"error" example:"Group does not exist"`
}

func (a *AccessRequest) Admit() *AccessRequest {
	a.Id = uuid.NewString()
	return a
}

func (a *AccessRequest) SetRequester(requester string) *AccessRequest {
	a.Status.RequestedBy = requester
	return a
}

// Method to approve the access request
func (a *AccessRequest) SetStatusApprove(approvedBy string) *AccessRequest {
	a.Status.Status = AccessRequestApproved
	a.Status.ApprovedBy = approvedBy
	return a
}

// Method to deny the access request
func (a *AccessRequest) SetStatusDenied(approvedBy string) *AccessRequest {
	a.Status.Status = AccessRequestDenied
	a.Status.ApprovedBy = approvedBy
	return a
}

// Method to expire the access request
func (a *AccessRequest) SetStatusExpired() *AccessRequest {
	a.Status.Status = AccessRequestExpired
	return a
}

// Method to set the access request to pending
func (a *AccessRequest) SetStatusPending() *AccessRequest {
	a.Status.Status = AccessRequestPending
	return a
}

func (s *AccessRequest) GetProviderUsername(provider string) string {
	if value, exists := s.Status.ProviderUsernames[provider]; exists {
		return value
	}
	return ""
}

func (s *AccessRequest) SetProviderUsernames(usernames map[string]string) *AccessRequest {

	if s.Status.ProviderUsernames == nil {
		s.Status.ProviderUsernames = make(map[string]string)
	}

	s.Status.ProviderUsernames = usernames
	return s
}

func (s *AccessRequest) SetProviderUsername(provider string, value string) *AccessRequest {

	if s.Status.ProviderUsernames == nil {
		s.Status.ProviderUsernames = make(map[string]string)
	}

	s.Status.ProviderUsernames[provider] = value
	return s
}

func (s *AccessRequest) SetApprovalRule(rule ApprovalRule) *AccessRequest {

	s.Status.ApprovalRule = rule
	return s
}

func (s *AccessRequest) GetApprovalRule() ApprovalRule {

	return s.Status.ApprovalRule
}

func (s *AccessRequest) GetRole(roles []AccessRole) (AccessRole, error) {

	for _, role := range roles {
		if s.RoleRef.Name == role.Name {
			return role, nil
		}
	}
	return AccessRole{}, fmt.Errorf("role not found: %s", s.RoleRef.Name)
}

func (s *AccessRequest) SetProviderStatusGranted(provider string, details string, err string) *AccessRequest {

	if s.Status.ProviderStatuses == nil {
		s.Status.ProviderStatuses = make(map[string]ProviderStatus)
	}

	s.Status.ProviderStatuses[provider] = ProviderStatus{
		Action:  ProviderStatusGranted,
		Details: details,
		Error:   err,
	}
	return s
}

func (s *AccessRequest) SetProviderStatusRevoked(provider string, details string, err string) *AccessRequest {

	if s.Status.ProviderStatuses == nil {
		s.Status.ProviderStatuses = make(map[string]ProviderStatus)
	}

	s.Status.ProviderStatuses[provider] = ProviderStatus{
		Action:  ProviderStatusRevoked,
		Details: details,
		Error:   err,
	}
	return s
}

func (s *AccessRequest) SetProviderStatusError(provider string, details string, err string) *AccessRequest {

	if s.Status.ProviderStatuses == nil {
		s.Status.ProviderStatuses = make(map[string]ProviderStatus)
	}

	s.Status.ProviderStatuses[provider] = ProviderStatus{
		Action:  ProviderStatusError,
		Details: details,
		Error:   err,
	}
	return s
}

func (s *AccessRequest) HasPermissions(user string, groups []string, utype string) bool {

	rule := s.Status.ApprovalRule

	// Check if it is a token
	if utype == "token" {
		return true
	}

	// Check if the user is explicitly listed
	if slices.Contains(rule.Users, user) {
		return true
	}

	// If no matching rule is found, deny approval
	if rule.Name == "" {
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

func (s *AccessRequest) SetTraceId(ctx context.Context) *AccessRequest {
	// ctx := context.Background() // Use your function's actual context here
	span := trace.SpanFromContext(ctx)

	s.Status.Trace = span.SpanContext().TraceID().String()

	return s
}

func (s *AccessRequest) SetExpiration(ctx context.Context) *AccessRequest {

	duration, _ := time.ParseDuration(s.Details.TTL)
	expires := time.Now().Add(duration)

	s.Status.ExpiresAt = &expires

	return s
}
