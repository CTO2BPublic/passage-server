package gitlab

import (
	"context"
	"errors"
	"fmt"

	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/rs/zerolog/log"
	clientgo "gitlab.com/gitlab-org/api/client-go"
	"go.opentelemetry.io/otel/attribute"
)

// Helper functions
func (a *GitlabProvider) getGroup(ctx context.Context) (group *clientgo.Group, err error) {

	_, span := tracing.NewSpanWrapper(ctx, "gitlab.getGroup")
	span.SetAttributes(
		attribute.String("peer.service", "gitlab"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	parameters := a.Parameters

	groups, _, err := a.Client.Groups.ListGroups(&clientgo.ListGroupsOptions{
		Search: &parameters.Group,
	})
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		log.Debug().Msgf("found gitlab group: %+v", group)
		if group.FullPath == parameters.Group {
			return group, nil
		}
	}
	return nil, fmt.Errorf("group not found")
}

func (a *GitlabProvider) getUser(ctx context.Context) (user *clientgo.User, err error) {

	_, span := tracing.NewSpanWrapper(ctx, "gitlab.getUser")
	span.SetAttributes(
		attribute.String("peer.service", "gitlab"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	parameters := a.Parameters

	users, _, err := a.Client.Users.ListUsers(&clientgo.ListUsersOptions{
		ListOptions: clientgo.ListOptions{},
		Search:      &a.Parameters.Username,
	})
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		log.Debug().Msgf("found user: %+v", user)
		if user.Username == parameters.Username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (a *GitlabProvider) addGroupMember(ctx context.Context, group *clientgo.Group, user *clientgo.User) (err error) {

	_, span := tracing.NewSpanWrapper(ctx, "gitlab.addGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "gitlab"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	parameters := a.Parameters
	level := clientgo.Ptr(parameters.Level)

	_, _, err = a.Client.GroupMembers.AddGroupMember(group.ID, &clientgo.AddGroupMemberOptions{
		Username:    &user.Username,
		AccessLevel: level,
		// ExpiresAt:    new(string),
	})
	return err
}

func (a *GitlabProvider) removeGroupMember(ctx context.Context, group *clientgo.Group, user *clientgo.User) (err error) {

	_, span := tracing.NewSpanWrapper(ctx, "gitlab.removeGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "gitlab"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	_, err = a.Client.GroupMembers.RemoveGroupMember(group.ID, user.ID, &clientgo.RemoveGroupMemberOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	return err
}

func (a *GitlabProvider) isGroupMember(ctx context.Context, group *clientgo.Group, user *clientgo.User) (isMember bool, err error) {

	_, span := tracing.NewSpanWrapper(ctx, "gitlab.isGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "gitlab"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	parameters := a.Parameters

	isMember = false
	users, _, err := a.Client.Groups.ListGroupMembers(group.ID, &clientgo.ListGroupMembersOptions{
		Query: &parameters.Username,
	})
	if err != nil {
		return isMember, fmt.Errorf("failed to retrieve users for group: %s: %w", parameters.Group, err)
	}

	for _, u := range users {
		log.Debug().Msgf("%+v", user)
		if u.Username == parameters.Username {
			isMember = true
			break
		}
	}
	return isMember, nil
}

// parseGitLabProviderConfig extracts the group ID, user ID, and token from the provider configuration
func extractParameters(config models.ProviderConfig) (GitlabProviderParameters, error) {

	data := config.Parameters

	creds := Config.GetCredentials(config.CredentialRef.Name)
	token := creds.GetString("token")

	group, ok := data["group"]
	if !ok {
		return GitlabProviderParameters{}, errors.New("group not found in provider config")
	}

	username, ok := data["username"]
	if !ok {
		return GitlabProviderParameters{}, errors.New("username not found in provider config")
	}

	level, ok := data["level"]
	if !ok {
		return GitlabProviderParameters{}, errors.New("level not found or empty in provider config")
	}

	var accessLevel clientgo.AccessLevelValue

	switch level {
	case "Owner":
		accessLevel = clientgo.OwnerPermissions
	case "Maintainer":
		accessLevel = clientgo.MaintainerPermissions
	case "Developer":
		accessLevel = clientgo.DeveloperPermissions
	case "Reporter":
		accessLevel = clientgo.ReporterPermissions
	case "Guest":
		accessLevel = clientgo.GuestPermissions
	}

	return GitlabProviderParameters{
		Group:    group,
		Token:    token,
		Username: username,
		Level:    accessLevel,
	}, nil
}

func StringPtr(s string) *string {
	return &s
}
