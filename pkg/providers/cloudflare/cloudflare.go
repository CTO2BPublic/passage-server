package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"
	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/iam"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/rs/zerolog/log"
)

const providerType = "cloudflare"

var Config = config.GetConfig()

type CloudflareProvider struct {
	client *cloudflare.Client

	// provider name as defined in the provider configuration
	name string

	// Cloudflare account ID
	accountID string

	// name of the cloudflare user group to manage access
	groupName string
	// ID of the cloudflare user group to manage access
	// we requires the ID because all the API calls need it
	groupID string
}

func NewCloudflareProvider(ctx context.Context, config models.ProviderConfig) (*CloudflareProvider, error) {
	creds := Config.GetCredentials(config.CredentialRef.Name)
	filePath := creds.GetString("credentialsfile")
	if filePath == "" {
		return nil, errors.New("credentialsfile not found in provider config")
	}
	apiToken, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read apiToken from file: %w", err)
	}

	accountID, ok := config.Parameters["accountID"]
	if !ok {
		return nil, errors.New("accountID not found in provider config")
	}

	groupName, ok := config.Parameters["group"]
	if !ok {
		return nil, errors.New("group not found in provider config")
	}

	client := cloudflare.NewClient(
		option.WithAPIToken(string(apiToken)),
	)

	p := &CloudflareProvider{
		client:    client,
		name:      config.Name,
		accountID: accountID,
		groupName: groupName,
	}

	group, err := p.findUserGroup(ctx, groupName)
	if err != nil {
		return nil, err
	}
	p.groupID = group.ID

	return p, nil
}

// GrantAccess adds an account member to a user group
func (p *CloudflareProvider) GrantAccess(ctx context.Context, request *models.AccessRequest) error {

	ctx, span := startSpan(ctx, "GrantAccess")
	defer span.End()

	username := request.GetProviderUsername(providerType)

	// Check if user is already a member
	isMember, err := p.isGroupMember(ctx, p.groupID, username)
	if err != nil {
		request.SetProviderStatusError(p.name, p.groupName, err.Error())
		return fmt.Errorf("failed to check group membership: %w", err)
	}

	if isMember {
		// User already in the group
		request.SetProviderStatusGranted(p.name, p.groupName, "already in group")
		log.Info().
			Str("TraceID", span.GetTraceID()).
			Str("Provider", p.name).
			Str("AccessRequest", request.Id).
			Str("Username", username).
			Str("Group", p.groupName).
			Msg("User already in group")
		return nil
	}

	err = p.addGroupMember(ctx, p.groupID, username)
	if err != nil {
		request.SetProviderStatusError(p.name, p.groupName, err.Error())
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	// User added successfully
	request.SetProviderStatusGranted(p.name, p.groupName, "")
	log.Info().
		Str("TraceID", span.GetTraceID()).
		Str("Provider", p.name).
		Str("AccessRequest", request.Id).
		Str("Username", username).
		Str("Group", p.groupName).
		Msg("User added to group")

	return nil
}

// RevokeAccess removes an account member from a user group
func (p *CloudflareProvider) RevokeAccess(ctx context.Context, request *models.AccessRequest) error {

	ctx, span := startSpan(ctx, "RevokeAccess")
	defer span.End()

	username := request.GetProviderUsername(providerType)

	// Check if user is already a member
	isMember, err := p.isGroupMember(ctx, p.groupID, username)
	if err != nil {
		request.SetProviderStatusError(p.name, p.groupName, err.Error())
		return fmt.Errorf("failed to check group membership: %w", err)
	}

	if !isMember {
		request.SetProviderStatusGranted(p.name, p.groupName, "already removed from group")
		log.Info().
			Str("TraceID", span.GetTraceID()).
			Str("Provider", p.name).
			Str("AccessRequest", request.Id).
			Str("Username", username).
			Str("Group", p.groupName).
			Msg("User already not in group")
		return nil
	}

	err = p.removeGroupMember(ctx, p.groupID, username)
	if err != nil {
		request.SetProviderStatusError(p.name, p.groupName, err.Error())
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	// User removed successfully
	request.SetProviderStatusRevoked(p.name, p.groupName, "")
	log.Info().
		Str("TraceID", span.GetTraceID()).
		Str("Provider", p.name).
		Str("AccessRequest", request.Id).
		Str("Username", username).
		Str("Group", p.groupName).
		Msg("User removed from group")

	return nil
}

func (p *CloudflareProvider) ListUsersWithAccess(ctx context.Context, role models.AccessRoleRef) ([]string, error) {

	ctx, span := startSpan(ctx, "ListUsersWithAccess")
	defer span.End()

	userEmails := []string{}

	page, err := p.client.IAM.UserGroups.Members.List(ctx, p.groupID, iam.UserGroupMemberListParams{
		AccountID: cloudflare.F(p.accountID),
	})

	for page != nil {
		for _, member := range page.Result {
			userEmails = append(userEmails, member.ID)
		}
		page, err = page.GetNextPage()
	}
	if err != nil {
		return nil, err
	}

	return userEmails, nil
}

func (p *CloudflareProvider) IsAccessExpired(ctx context.Context, request *models.AccessRequest) (bool, error) {

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
	ctx, span := tracing.NewSpanWrapper(ctx, fmt.Sprintf("providers.cloudflare.%s", name))
	return ctx, span
}
