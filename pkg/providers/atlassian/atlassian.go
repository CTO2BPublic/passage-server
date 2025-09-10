package atlassian

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"
	jira "github.com/ctreminiom/go-atlassian/v2/jira/v2"
	"go.opentelemetry.io/otel"

	"github.com/rs/zerolog/log"
)

const providerType = "cloudflare"

var Config = config.GetConfig()
var Tracer = otel.Tracer("pkg/providers/atlassian")

// AtlassianProvider handles Atlassian access management
type AtlassianProvider struct {
	client    *jira.Client
	accountID string
	groupName string
	name      string
}

// NewAtlassianProvider initializes a new AtlassianProvider with credentials from ProviderConfig
func NewAtlassianProvider(ctx context.Context, config models.ProviderConfig) (*AtlassianProvider, error) {
	creds := Config.GetCredentials(config.CredentialRef.Name)
	apiToken := creds.GetString("token")
	if apiToken == "" {
		return nil, errors.New("token not found in provider config")
	}
	apiToken = strings.TrimSpace(apiToken)

	email := creds.GetString("email")
	if email == "" {
		return nil, fmt.Errorf("credential email is missing")
	}

	data := config.Parameters
	group, ok := data["group"]
	if !ok {
		return nil, errors.New("group not found in provider config")
	}
	siteURL := data["siteurl"]
	if siteURL == "" {
		return nil, errors.New("siteurl not found in provider config")
	}

	client, err := jira.New(
		http.DefaultClient,
		siteURL,
	)
	if err != nil {
		return nil, err
	}
	client.Auth.SetBasicAuth(email, string(apiToken))

	return &AtlassianProvider{
		client:    client,
		groupName: group,
		name:      config.Name,
	}, nil
}

// GrantAccess adds a user to a specified group based on the provider parameters
func (a *AtlassianProvider) GrantAccess(ctx context.Context, request *models.AccessRequest) error {
	ctx, span := startSpan(ctx, "GrantAccess")
	defer span.End()

	username := request.GetProviderUsername(providerType)

	// Check if user is already a member
	isMember, err := a.isGroupMember(ctx, a.groupName, username)
	if err != nil {
		request.SetProviderStatusError(a.name, a.groupName, err.Error())
		return fmt.Errorf("failed to check group membership: %w", err)
	}

	if isMember {
		// User already in the group
		request.SetProviderStatusGranted(a.name, a.groupName, "already in group")
		log.Info().
			Str("TraceID", span.GetTraceID()).
			Str("Provider", a.name).
			Str("AccessRequest", request.Id).
			Str("Username", username).
			Str("Group", a.groupName).
			Msg("User already in group")
		return nil
	}

	err = a.addGroupMember(ctx, a.groupName, username)
	if err != nil {
		request.SetProviderStatusError(a.name, a.groupName, err.Error())
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	// User added successfully
	request.SetProviderStatusGranted(a.name, a.groupName, "")
	log.Info().
		Str("TraceID", span.GetTraceID()).
		Str("Provider", a.name).
		Str("AccessRequest", request.Id).
		Str("Username", username).
		Str("Group", a.groupName).
		Msg("User added to group")

	return nil
}

// RevokeAccess removes a user from a specified group based on the provider parameters
func (a *AtlassianProvider) RevokeAccess(ctx context.Context, request *models.AccessRequest) error {
	ctx, span := startSpan(ctx, "RevokeAccess")
	defer span.End()

	username := request.GetProviderUsername(providerType)

	// Check if user is already a member
	isMember, err := a.isGroupMember(ctx, a.groupName, username)
	if err != nil {
		request.SetProviderStatusError(a.name, a.groupName, err.Error())
		return fmt.Errorf("failed to check group membership: %w", err)
	}

	if !isMember {
		request.SetProviderStatusGranted(a.name, a.groupName, "already removed from group")
		log.Info().
			Str("TraceID", span.GetTraceID()).
			Str("Provider", a.name).
			Str("AccessRequest", request.Id).
			Str("Username", username).
			Str("Group", a.groupName).
			Msg("User already not in group")
		return nil
	}

	err = a.removeGroupMember(ctx, a.groupName, username)
	if err != nil {
		request.SetProviderStatusError(a.name, a.groupName, err.Error())
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	// User removed successfully
	request.SetProviderStatusRevoked(a.name, a.groupName, "")
	log.Info().
		Str("TraceID", span.GetTraceID()).
		Str("Provider", a.name).
		Str("AccessRequest", request.Id).
		Str("Username", username).
		Str("Group", a.groupName).
		Msg("User removed from group")

	return nil
}

// ListUsersWithAccess lists users with access to the specified role
func (a *AtlassianProvider) ListUsersWithAccess(ctx context.Context, roleRef models.AccessRoleRef) ([]string, error) {
	ctx, span := startSpan(ctx, "ListUsersWithAccess")
	defer span.End()

	members, err := a.getGroupMembers(ctx, roleRef.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list group members: %w", err)
	}

	usernames := []string{}
	for _, member := range members {
		usernames = append(usernames, member.EmailAddress)
	}

	return usernames, nil
}

// IsAccessExpired checks whether the access for the given request has expired
func (a *AtlassianProvider) IsAccessExpired(ctx context.Context, request *models.AccessRequest) (bool, error) {
	_, span := startSpan(ctx, "IsAccessExpired")
	defer span.End()

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

func startSpan(ctx context.Context, name string) (context.Context, *tracing.SpanWrapper) {
	ctx, span := tracing.NewSpanWrapper(ctx, fmt.Sprintf("providers.atlassian.%s", name))
	return ctx, span
}
