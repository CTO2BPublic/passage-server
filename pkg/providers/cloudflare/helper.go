package cloudflare

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
	"github.com/cloudflare/cloudflare-go/v6/iam"
	"github.com/cloudflare/cloudflare-go/v6/shared"
	"go.opentelemetry.io/otel/attribute"
)

var (
	errMemberNotFound    = errors.New("member not found")
	errUserGroupNotFound = errors.New("user group not found")
)

// findUserGroup retrieves a user group by name
func (p *CloudflareProvider) findUserGroup(ctx context.Context, groupName string) (iam.UserGroupListResponse, error) {
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

func (p *CloudflareProvider) findMember(ctx context.Context, email string) (shared.Member, error) {
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

func (p *CloudflareProvider) addGroupMember(ctx context.Context, groupID string, username string) error {
	ctx, span := startSpan(ctx, "addGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "cloudflare"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	member, err := p.findMember(ctx, username)
	if err != nil {
		return err
	}

	_, err = p.client.IAM.UserGroups.Members.New(ctx, groupID, iam.UserGroupMemberNewParams{
		AccountID: cloudflare.F(p.accountID),
		Body: []iam.UserGroupMemberNewParamsBody{
			{ID: cloudflare.F(member.ID)},
		},
	})

	return err
}

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
