package google

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

var Config = config.GetConfig()
var Tracer = otel.Tracer("pkg/providers/google")

// GoogleProvider handles Google Workspace access management
type GoogleProvider struct {
	Service    *admin.Service
	Parameters GoogleProviderParameters
	Name       string `json:"name"`
}

// GoogleProviderParameters encapsulates extracted provider details
type GoogleProviderParameters struct {
	Group           string `json:"group"`
	Username        string `json:"username"`
	CredentialsFile string `json:"credentialsFIle"`
}

// NewGoogleProvider initializes a new GoogleProvider with credentials from ProviderConfig
func NewGoogleProvider(ctx context.Context, config models.ProviderConfig) (*GoogleProvider, error) {
	parameters, err := extractParameters(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider config: %w", err)
	}

	// Initialize the Admin SDK Directory API client
	service, err := admin.NewService(
		context.Background(),
		option.WithCredentialsFile(parameters.CredentialsFile),
		option.WithScopes(
			admin.AdminDirectoryGroupScope,
			admin.AdminDirectoryGroupMemberScope,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Admin SDK client: %w", err)
	}

	return &GoogleProvider{Service: service, Parameters: parameters, Name: config.Name}, nil
}

// GrantAccess adds a user to a specified Google Workspace group
func (g *GoogleProvider) GrantAccess(ctx context.Context, request *models.AccessRequest) error {

	ctx, span := tracing.NewSpanWrapper(ctx, "providers.google.GrantAccess")
	defer span.End()

	parameters := g.Parameters

	// Check if user is already a member
	isMember, err := g.isGroupMember(ctx, parameters.Group, parameters.Username)
	if err != nil {
		request.SetProviderStatusError(g.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to check group membership: %w", err)
	}

	if isMember {
		// User already in the group
		request.SetProviderStatusGranted(g.Name, parameters.Group, "already in group")
		log.Info().
			Str("TraceID", span.GetTraceID()).
			Str("Provider", g.Name).
			Str("AccessRequest", request.Id).
			Str("Username", parameters.Username).
			Str("Group", parameters.Group).
			Msg("User already in group")
		return nil
	}

	err = g.addGroupMember(ctx, parameters.Group, parameters.Username)
	if err != nil {
		request.SetProviderStatusError(g.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	// User added successfully
	request.SetProviderStatusGranted(g.Name, parameters.Group, "")
	log.Info().
		Str("TraceID", span.GetTraceID()).
		Str("Provider", g.Name).
		Str("AccessRequest", request.Id).
		Str("Username", parameters.Username).
		Str("Group", parameters.Group).
		Msg("User added to group")

	return nil
}

// RevokeAccess removes a user from a specified Google Workspace group
func (g *GoogleProvider) RevokeAccess(ctx context.Context, request *models.AccessRequest) error {

	ctx, span := tracing.NewSpanWrapper(ctx, "providers.google.RevokeAccess")
	defer span.End()

	parameters := g.Parameters

	// Check if the user is already not in the group
	isMember, err := g.isGroupMember(ctx, parameters.Group, parameters.Username)
	if err != nil {
		request.SetProviderStatusError(g.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to check group membership: %w", err)
	}

	if !isMember {
		request.SetProviderStatusRevoked(g.Name, parameters.Group, "already removed from group")
		log.Info().
			Str("TraceID", span.GetTraceID()).
			Str("Provider", g.Name).
			Str("AccessRequest", request.Id).
			Str("Username", parameters.Username).
			Str("Group", parameters.Group).
			Msg("User already not in group")
		return nil
	}

	// Remove user from group
	err = g.removeGroupMember(ctx, parameters.Group, parameters.Username)
	if err != nil {
		request.SetProviderStatusError(g.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	// User removed successfully
	request.SetProviderStatusRevoked(g.Name, parameters.Group, "")
	log.Info().
		Str("TraceID", span.GetTraceID()).
		Str("Provider", g.Name).
		Str("AccessRequest", request.Id).
		Str("Username", parameters.Username).
		Str("Group", parameters.Group).
		Msg("User removed from group")

	return nil
}

// ListUsersWithAccess lists all users in a specified Google Workspace group
func (g *GoogleProvider) ListUsersWithAccess(ctx context.Context, roleRef models.AccessRoleRef) ([]string, error) {

	_, span := tracing.NewSpanWrapper(ctx, "providers.google.ListUsersWithAccess")
	defer span.End()

	parameters := g.Parameters

	members, err := g.Service.Members.List(parameters.Group).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list group members: %w", err)
	}

	userEmails := []string{}
	for _, member := range members.Members {
		userEmails = append(userEmails, member.Email)
	}

	return userEmails, nil
}

// IsAccessExpired checks whether the access for the given request has expired
func (g *GoogleProvider) IsAccessExpired(ctx context.Context, request *models.AccessRequest) (bool, error) {
	ttl := request.Details.TTL
	if ttl == "" {
		return false, errors.New("TTL not specified in access request")
	}

	// Validate TTL expiration (this assumes TTL is a duration like "24h")
	expiry, err := time.ParseDuration(ttl)
	if err != nil {
		return false, fmt.Errorf("invalid TTL format: %w", err)
	}

	expirationTime := request.CreatedAt.Add(expiry)
	return time.Now().After(expirationTime), nil
}
