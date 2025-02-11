package gitlab

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/rs/zerolog/log"
	clientgo "gitlab.com/gitlab-org/api/client-go"
	"go.opentelemetry.io/otel"
)

var Config = config.GetConfig()
var Tracer = otel.Tracer("pkg/providers/gitlab")

// GitlabProvider handles GitLab access management
// Credentials and configurations are specified in the ProviderConfig parameters
type GitlabProvider struct {
	Client     *clientgo.Client
	Parameters GitlabProviderParameters
	Name       string `json:"name"`
}

// parameters encapsulates extracted provider details
type GitlabProviderParameters struct {
	Group    string                    `json:"group"`
	Username string                    `json:"username"`
	Token    string                    `json:"token"`
	Level    clientgo.AccessLevelValue `json:"string"`
}

// NewGitlabProvider initializes a new GitlabProvider with credentials from ProviderConfig
func NewGitlabProvider(ctx context.Context, config models.ProviderConfig) (*GitlabProvider, error) {
	parameters, err := extractParameters(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider config: %w", err)
	}

	client, err := clientgo.NewClient(parameters.Token, clientgo.WithBaseURL("https://gitlab.com/api/v4"))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &GitlabProvider{Client: client, Parameters: parameters, Name: config.Name}, nil
}

// GrantAccess adds a user to a specified group based on the provider parameters
func (a *GitlabProvider) GrantAccess(ctx context.Context, request *models.AccessRequest) error {

	ctx, span := tracing.NewSpanWrapper(ctx, "providers.gitlab.GrantAccess")
	defer span.End()

	parameters := a.Parameters

	group, err := a.getGroup(ctx)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to retrieve gitlab group %s: %w", parameters.Group, err)
	}

	user, err := a.getUser(ctx)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to retrieve gitlab user: %s: %w", parameters.Username, err)
	}

	err = a.addGroupMember(ctx, group, user)
	if err != nil {
		if apiErr, ok := err.(*clientgo.ErrorResponse); ok && apiErr.Response.StatusCode == 409 {

			// Member already exists. Update request status
			request.SetProviderStatusGranted(a.Name, parameters.Group, "already in group")
			log.Info().
				Str("Provider", a.Name).
				Str("AccessRequest", request.Id).
				Str("username", parameters.Username).
				Str("group", parameters.Group).
				Msg("User already in group")
			return nil
		}

		// Failed to add user. Update request status
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())

		return fmt.Errorf("failed to add user: %s to group: %s: %w", parameters.Username, parameters.Group, err)
	}

	// User added
	request.SetProviderStatusGranted(a.Name, parameters.Group, "")
	log.Info().
		Str("Provider", a.Name).
		Str("AccessRequest", request.Id).
		Str("username", parameters.Username).
		Str("group", parameters.Group).
		Msg("User added to group")

	return nil
}

// RevokeAccess removes a user from a specified group based on the provider parameters
func (a *GitlabProvider) RevokeAccess(ctx context.Context, request *models.AccessRequest) error {

	ctx, span := tracing.NewSpanWrapper(ctx, "providers.gitlab.RevokeAccess")
	defer span.End()

	parameters := a.Parameters

	group, err := a.getGroup(ctx)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to retrieve gitlab group: %w", err)
	}

	user, err := a.getUser(ctx)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to retrieve gitlab user id: %w", err)
	}

	isMember, err := a.isGroupMember(ctx, group, user)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to check user membership: %w", err)
	}

	if !isMember {
		request.SetProviderStatusRevoked(a.Name, parameters.Group, "already removed from group")
		log.Info().
			Str("Provider", a.Name).
			Str("AccessRequest", request.Id).
			Str("username", parameters.Username).
			Str("group", parameters.Group).
			Msg("User already not in group")
		return nil
	}

	err = a.removeGroupMember(ctx, group, user)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	request.SetProviderStatusRevoked(a.Name, parameters.Group, "")
	log.Info().
		Str("Provider", a.Name).
		Str("AccessRequest", request.Id).
		Str("username", parameters.Username).
		Str("group", parameters.Group).
		Msgf("User removed from group")
	return nil

}

// ListUsersWithAccess lists users with access to the specified role
func (a *GitlabProvider) ListUsersWithAccess(ctx context.Context, roleRef models.AccessRoleRef) ([]string, error) {

	ctx, span := tracing.NewSpanWrapper(ctx, "providers.gitlab.ListUsersWithAccess")
	defer span.End()

	parameters := a.Parameters

	group, err := a.getGroup(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse group ID: %w", err)
	}

	members, _, err := a.Client.Search.UsersByGroup(group, parameters.Group, &clientgo.SearchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list group members: %w", err)
	}

	usernames := []string{}
	for _, member := range members {
		usernames = append(usernames, member.Username)
	}

	return usernames, nil
}

// IsAccessExpired checks whether the access for the given request has expired
func (a *GitlabProvider) IsAccessExpired(ctx context.Context, request *models.AccessRequest) (bool, error) {
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
