package atlassian

import (
	"context"
	"errors"
	"fmt"

	"github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"

	"go.opentelemetry.io/otel/attribute"
)

var (
	errMemberNotFound = errors.New("member not found")
)

func (a *AtlassianProvider) getUser(ctx context.Context, username string) (*models.UserScheme, error) {
	_, span := startSpan(ctx, "getUser")
	span.SetAttributes(
		attribute.String("peer.service", "atlassian"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	var (
		startAt    = 0
		maxResults = 1000 // FIXME: this will be a problem if we have more than 1000 users
	)

	users, _, err := a.client.User.Gets(ctx, startAt, maxResults)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found: %s", a.accountID)
	}

	for _, user := range users {
		if user.Name == username || user.EmailAddress == username {
			return user, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", errMemberNotFound, username)
}

func (a *AtlassianProvider) addGroupMember(ctx context.Context, groupName string, username string) error {
	_, span := startSpan(ctx, "addGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "atlassian"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	user, err := a.getUser(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if _, resp, err := a.client.Group.Add(ctx, groupName, user.AccountID); err != nil {
		return fmt.Errorf("failed to add user to group: %s", resp.Bytes.String())
	}

	return nil
}

func (a *AtlassianProvider) removeGroupMember(ctx context.Context, groupName string, username string) error {
	_, span := startSpan(ctx, "removeGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "atlassian"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	user, err := a.getUser(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if _, err := a.client.Group.Remove(ctx, groupName, user.AccountID); err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}
	return nil
}

func (a *AtlassianProvider) isGroupMember(ctx context.Context, groupName string, username string) (bool, error) {
	_, span := startSpan(ctx, "isGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "atlassian"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	user, err := a.getUser(ctx, username)
	if err != nil {
		return false, fmt.Errorf("failed to find user: %w", err)
	}

	groups, _, err := a.client.User.Groups(ctx, user.AccountID)
	if err != nil {
		return false, err
	}

	for _, group := range groups {
		if group.Name == groupName {
			return true, nil
		}
	}

	return false, nil
}

func (a *AtlassianProvider) getGroupMembers(ctx context.Context, groupName string) ([]*models.GroupUserDetailScheme, error) {
	_, span := startSpan(ctx, "getGroupMembers")
	span.SetAttributes(
		attribute.String("peer.service", "atlassian"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	var (
		inactive   = false
		startAt    = 0
		maxResults = 100
		allMembers = make([]*models.GroupUserDetailScheme, 0, maxResults)
	)

	members, _, err := a.client.Group.Members(ctx, groupName, inactive, startAt, maxResults)
	if err != nil {
		return nil, err
	}

	for members != nil {
		allMembers = append(allMembers, members.Values...)

		if members.IsLast {
			break
		}

		startAt += members.MaxResults
		members, _, err = a.client.Group.Members(ctx, groupName, inactive, startAt, maxResults)
		if err != nil {
			return nil, err
		}
	}

	return allMembers, nil
}
