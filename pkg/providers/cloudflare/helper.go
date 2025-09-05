package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
	"github.com/cloudflare/cloudflare-go/v6/iam"
	"github.com/cloudflare/cloudflare-go/v6/shared"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"
)

var (
	errMemberNotFound    = errors.New("member not found")
	errUserGroupNotFound = errors.New("user group not found")
)

// findUserGroup retrieves a user group by name
func (p *CloudflareProvider) findUserGroup(ctx context.Context, groupName string) (iam.UserGroupListResponse, error) {
	ctx, span := startSpan(ctx, "findUserGroup")
	span.SetAttributes(
		attribute.String("peer.service", "cloudflare"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	iter := p.client.IAM.UserGroups.ListAutoPaging(ctx, iam.UserGroupListParams{
		AccountID: cloudflare.F(p.accountID),
		Name:      cloudflare.F(p.groupName),
	})

	for iter.Next() {
		group := iter.Current()
		if group.Name == groupName {
			return group, nil
		}
	}
	if err := iter.Err(); err != nil {
		return iam.UserGroupListResponse{}, fmt.Errorf("failed to list user groups: %w", err)
	}

	return iam.UserGroupListResponse{}, fmt.Errorf("%w: %s", errUserGroupNotFound, groupName)
}

// findMember retrieves an account member by email
func (p *CloudflareProvider) findMember(ctx context.Context, email string) (shared.Member, error) {
	ctx, span := startSpan(ctx, "findMember")
	span.SetAttributes(
		attribute.String("peer.service", "cloudflare"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	iter := p.client.Accounts.Members.ListAutoPaging(ctx, accounts.MemberListParams{
		AccountID: cloudflare.F(p.accountID),
	})

	for iter.Next() {
		member := iter.Current()
		if member.Email == email {
			return member, nil
		}
	}
	if err := iter.Err(); err != nil {
		return shared.Member{}, fmt.Errorf("failed to list account members: %w", err)
	}

	return shared.Member{}, fmt.Errorf("%w: %s", errMemberNotFound, email)
}

// addAccountMember adds a new member to the account of the provider
// the member is invited with minimal account access possible
func (p *CloudflareProvider) addAccountMember(ctx context.Context, email string) error {
	ctx, span := startSpan(ctx, "addAccountMember")
	span.SetAttributes(
		attribute.String("peer.service", "cloudflare"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	// when we invite a user, we give it the minimal account access
	// the rest of the permission will be given by the user group he is added to
	const roleName = "Minimal Account Access"
	var roleID string

	iter := p.client.Accounts.Roles.ListAutoPaging(ctx, accounts.RoleListParams{
		AccountID: cloudflare.F(p.accountID),
	})
	for iter.Next() {
		role := iter.Current()
		if role.Name == roleName {
			roleID = role.ID
			break
		}
	}
	if iter.Err() != nil {
		return fmt.Errorf("failed to list account roles: %w", iter.Err())
	}

	if roleID == "" {
		return fmt.Errorf("role %q not found in account %s", roleName, p.accountID)
	}

	_, err := p.client.Accounts.Members.New(ctx, accounts.MemberNewParams{
		AccountID: cloudflare.F(p.accountID),
		Body: accounts.MemberNewParamsBodyIAMCreateMemberWithRoles{
			Email: cloudflare.F(email),
			Roles: cloudflare.F([]string{roleID}),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to invite member %s to account %s: %w", email, p.accountID, err)
	}

	return nil
}

// removeAccountMember removes a member from the account of the provider
func (p *CloudflareProvider) removeAccountMember(ctx context.Context, email string) error {
	ctx, span := startSpan(ctx, "removeAccountMember")
	span.SetAttributes(
		attribute.String("peer.service", "cloudflare"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	member, err := p.findMember(ctx, email)
	if err != nil {
		// If the member is not found, nothing else to do
		if errors.Is(err, errMemberNotFound) {
			return nil
		}
		return err
	}

	_, err = p.client.Accounts.Members.Delete(ctx, member.ID, accounts.MemberDeleteParams{
		AccountID: cloudflare.F(p.accountID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete member %s from account %s: %w", email, p.accountID, err)
	}

	return nil
}

// addGroupMember adds an account member to a user group
func (p *CloudflareProvider) addGroupMember(ctx context.Context, groupID string, username string) error {
	ctx, span := startSpan(ctx, "addGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "cloudflare"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	member, err := p.findMember(ctx, username)
	if err != nil && !errors.Is(err, errMemberNotFound) {
		return err
	}

	if errors.Is(err, errMemberNotFound) {
		if err := p.addAccountMember(ctx, username); err != nil {
			return err
		}
	}

	_, err = p.client.IAM.UserGroups.Members.New(ctx, groupID, iam.UserGroupMemberNewParams{
		AccountID: cloudflare.F(p.accountID),
		Body: []iam.UserGroupMemberNewParamsBody{
			{ID: cloudflare.F(member.ID)},
		},
	})

	return err
}

// removeGroupMember removes an account member from a user group
func (p *CloudflareProvider) removeGroupMember(ctx context.Context, groupID string, username string) error {
	ctx, span := startSpan(ctx, "removeGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "cloudflare"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	member, err := p.findMember(ctx, username)
	if err != nil {
		// If the member is not found, nothing else to do
		if errors.Is(err, errMemberNotFound) {
			return nil
		}
		return err
	}

	_, err = p.client.IAM.UserGroups.Members.Delete(ctx, groupID, member.ID, iam.UserGroupMemberDeleteParams{
		AccountID: cloudflare.F(p.accountID),
	})

	return err
}

// isGroupMember checks if an account member is part of a user group
func (p *CloudflareProvider) isGroupMember(ctx context.Context, groupID, username string) (bool, error) {

	ctx, span := startSpan(ctx, "isGroupMember")
	defer span.End()
	span.SetAttributes(
		attribute.String("peer.service", "cloudflare"),
		attribute.String("span.kind", "client"),
	)

	iter := p.client.IAM.UserGroups.Members.ListAutoPaging(ctx, groupID, iam.UserGroupMemberListParams{
		AccountID: cloudflare.F(p.accountID),
	})

	for iter.Next() {
		member := iter.Current()
		if member.Email == username {
			return true, nil
		}
	}
	if err := iter.Err(); err != nil {
		return false, fmt.Errorf("failed to list account members: %w", err)
	}

	return false, nil
}

// memberGroups retrieves all groups a member belongs to concurrently
func (p *CloudflareProvider) memberGroups(ctx context.Context) (map[string][]iam.UserGroupListResponse, error) {
	groups := []iam.UserGroupListResponse{}

	iter := p.client.IAM.UserGroups.ListAutoPaging(ctx, iam.UserGroupListParams{
		AccountID: cloudflare.F(p.accountID),
	})
	for iter.Next() {
		group := iter.Current()
		groups = append(groups, group)
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list user groups: %w", err)
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.NumCPU())

	membersGroup := make(map[string][]iam.UserGroupListResponse)
	type memberGroup struct {
		Email string
		Group iam.UserGroupListResponse
	}
	c := make(chan memberGroup, 1)

	go func() {
		for mg := range c {
			_, exists := membersGroup[mg.Email]
			if !exists {
				membersGroup[mg.Email] = []iam.UserGroupListResponse{}
			}
			membersGroup[mg.Email] = append(membersGroup[mg.Email], mg.Group)
		}
	}()

	for _, group := range groups {
		g.Go(func() error {
			iter := p.client.IAM.UserGroups.Members.ListAutoPaging(gctx, group.ID, iam.UserGroupMemberListParams{
				AccountID: cloudflare.F(p.accountID),
			})

			for iter.Next() {
				member := iter.Current()
				c <- memberGroup{
					Email: member.Email,
					Group: group,
				}
			}

			if err := iter.Err(); err != nil {
				return fmt.Errorf("failed to list account members: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	close(c)

	return membersGroup, nil
}

// canMemberBeRemoved checks if a member can be removed from the account
// a member can be removed if it is not part of any user group anymore
func (p *CloudflareProvider) canMemberBeRemoved(ctx context.Context, username string) (bool, error) {

	memberGroups, err := p.memberGroups(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list member groups: %w", err)
	}

	_, exists := memberGroups[username]
	if !exists {
		// user is not in any group, it can be removed
		return true, nil
	}

	return false, nil
}
