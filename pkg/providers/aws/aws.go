package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var Config = config.GetConfig()
var Tracer = otel.Tracer("pkg/providers/aws")

// AWSProvider handles AWS Identity Center (SSO) group management
type AWSProvider struct {
	SSOAdminClient *ssoadmin.Client
	IdentityClient *identitystore.Client
	Parameters     AWSProviderParameters
	Name           string `json:"name"`
}

// AWSProviderParameters encapsulates provider configuration details
type AWSProviderParameters struct {
	IdentityStoreID string `json:"identitystoreid"`
	InstanceARN     string `json:"instancearn"`
	Region          string `json:"region"`
	Group           string `json:"group"`
	Username        string `json:"username"`
}

// NewAWSProvider initializes an AWSProvider with the given configuration
func NewAWSProvider(ctx context.Context, config models.ProviderConfig) (*AWSProvider, error) {
	parameters, err := extractParameters(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider config: %w", err)
	}

	creds := Config.GetCredentials(config.CredentialRef.Name)
	accessKeyId := creds.GetString("accesskeyid")
	secretAccessKey := creds.GetString("secretaccesskey")

	log.Info().Msg(parameters.Region)
	// Load AWS configuration
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(parameters.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK configuration: %w", err)
	}

	return &AWSProvider{
		SSOAdminClient: ssoadmin.NewFromConfig(awsConfig),
		IdentityClient: identitystore.NewFromConfig(awsConfig),
		Parameters:     parameters,
		Name:           config.Name,
	}, nil
}

// GrantAccess adds a user to an Identity Center group
func (a *AWSProvider) GrantAccess(ctx context.Context, request *models.AccessRequest) error {
	ctx, span := tracing.NewSpanWrapper(ctx, "providers.aws.GrantAccess")
	defer span.End()

	parameters := a.Parameters

	// Get group ID
	group, err := a.getGroup(ctx, parameters.Group)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return err
	}

	// Get user ID
	userID, err := a.getUser(ctx, parameters.Username)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return err
	}

	// Add user to group
	_, err = a.IdentityClient.CreateGroupMembership(ctx, &identitystore.CreateGroupMembershipInput{
		IdentityStoreId: aws.String(parameters.IdentityStoreID),
		GroupId:         group.GroupId,
		MemberId: &types.MemberIdMemberUserId{
			Value: *userID.UserId,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			request.SetProviderStatusGranted(a.Name, parameters.Group, err.Error())
			return err
		}
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return err
	}

	request.SetProviderStatusGranted(a.Name, parameters.Group, "")
	log.Info().
		Str("TraceID", span.GetTraceID()).
		Str("Provider", a.Name).
		Str("AccessRequest", request.Id).
		Str("Username", parameters.Username).
		Str("Group", parameters.Group).
		Msg("User added to group")
	return nil
}

// RevokeAccess removes a user from an Identity Center group
func (a *AWSProvider) RevokeAccess(ctx context.Context, request *models.AccessRequest) error {
	ctx, span := tracing.NewSpanWrapper(ctx, "providers.aws.RevokeAccess")
	defer span.End()

	parameters := a.Parameters

	// Get group ID
	group, err := a.getGroup(ctx, parameters.Group)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return err
	}

	// Get user ID
	user, err := a.getUser(ctx, parameters.Username)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return err
	}

	// Find membership ID
	membershipID, err := a.getMembershipID(ctx, group.GroupId, user.UserId)
	if err != nil {
		if strings.Contains(err.Error(), "membership not found") {
			// User is already removed from the group, set status and log it
			request.SetProviderStatusRevoked(a.Name, parameters.Group, "already removed from group")
			log.Info().
				Str("TraceID", span.GetTraceID()).
				Str("Provider", a.Name).
				Str("AccessRequest", request.Id).
				Str("Username", parameters.Username).
				Str("Group", parameters.Group).
				Msg("User already removed from group")
			return nil
		}
		request.SetProviderStatusError(a.Name, a.Parameters.Group, err.Error())
		return err
	}

	// Remove user from group
	_, err = a.IdentityClient.DeleteGroupMembership(ctx, &identitystore.DeleteGroupMembershipInput{
		IdentityStoreId: aws.String(parameters.IdentityStoreID),
		MembershipId:    aws.String(membershipID),
	})
	if err != nil {
		request.SetProviderStatusError(a.Name, a.Parameters.Group, err.Error())
		return nil
	}

	request.SetProviderStatusRevoked(a.Name, parameters.Group, "")
	log.Info().
		Str("TraceID", span.GetTraceID()).
		Str("Provider", a.Name).
		Str("AccessRequest", request.Id).
		Str("Username", parameters.Username).
		Str("Group", parameters.Group).
		Msg("User removed from group")
	return nil
}

// IsAccessExpired checks whether the access for the given request has expired
func (a *AWSProvider) IsAccessExpired(ctx context.Context, request *models.AccessRequest) (bool, error) {
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

func (a *AWSProvider) ListUsersWithAccess(ctx context.Context, roleRef models.AccessRoleRef) ([]string, error) {
	// Create a tracing span
	_, span := tracing.NewSpanWrapper(ctx, "providers.aws.ListUsersWithAccess")
	defer span.End()

	// Input for AWS Identity Store ListGroupMemberships API call
	input := &identitystore.ListGroupMembershipsInput{
		IdentityStoreId: aws.String(a.Parameters.IdentityStoreID),
		GroupId:         aws.String(roleRef.Name), // Use roleRef.RoleName or another appropriate value
	}

	// List group memberships using AWS SDK Identity Store client
	output, err := a.IdentityClient.ListGroupMemberships(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list group memberships for role %s: %w", roleRef.Name, err)
	}

	// Collect the user IDs (or other identifiers) from the group memberships
	var userEmails []string
	for _, membership := range output.GroupMemberships {
		if membership.MemberId != nil {
			// Type assertion to get UserId and assuming MemberId is of type MemberIdMemberUserId
			if memberIdUser, ok := membership.MemberId.(*types.MemberIdMemberUserId); ok {
				// Assuming you have a method to get user information (like email) using the UserId
				userEmails = append(userEmails, memberIdUser.Value)
			}
		}
	}

	// Return the list of user emails or IDs
	return userEmails, nil
}

// getGroupID retrieves the ID of a group by name
func (a *AWSProvider) getGroup(ctx context.Context, groupName string) (*types.Group, error) {

	ctx, span := tracing.NewSpanWrapper(ctx, "aws.getGroup")
	span.SetAttributes(
		attribute.String("peer.service", "aws"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	output, err := a.IdentityClient.ListGroups(ctx, &identitystore.ListGroupsInput{
		IdentityStoreId: aws.String(a.Parameters.IdentityStoreID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	for _, group := range output.Groups {
		if group.DisplayName != nil && *group.DisplayName == groupName {
			return &group, nil
		}
	}
	return nil, fmt.Errorf("group not found: %s", groupName)
}

// getUserID retrieves the ID of a user by username
func (a *AWSProvider) getUser(ctx context.Context, username string) (*types.User, error) {

	ctx, span := tracing.NewSpanWrapper(ctx, "aws.getUser")
	span.SetAttributes(
		attribute.String("peer.service", "aws"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	output, err := a.IdentityClient.ListUsers(ctx, &identitystore.ListUsersInput{
		IdentityStoreId: aws.String(a.Parameters.IdentityStoreID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	for _, user := range output.Users {
		if user.UserName != nil && *user.UserName == username {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found: %s", username)
}

// getMembershipID retrieves the ID of a group membership
func (a *AWSProvider) getMembershipID(ctx context.Context, groupID, userID *string) (string, error) {

	ctx, span := tracing.NewSpanWrapper(ctx, "aws.getMembershipID")
	span.SetAttributes(
		attribute.String("peer.service", "aws"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	// List group memberships
	output, err := a.IdentityClient.ListGroupMemberships(ctx, &identitystore.ListGroupMembershipsInput{
		IdentityStoreId: aws.String(a.Parameters.IdentityStoreID),
		GroupId:         groupID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list group memberships: %w", err)
	}

	// Iterate through the group memberships
	for _, membership := range output.GroupMemberships {
		// Check if MemberId is not nil and type assert to the correct type
		if membership.MemberId != nil {
			if memberIdUser, ok := membership.MemberId.(*types.MemberIdMemberUserId); ok {
				if memberIdUser.Value == *userID {
					return *membership.MembershipId, nil
				}
			}
		}
	}

	// Return error if membership is not found
	return "", fmt.Errorf("membership not found for group %s and user %s", *groupID, *userID)
}

// extractParameters extracts AWS provider parameters from the provider config
func extractParameters(config models.ProviderConfig) (AWSProviderParameters, error) {
	data := config.Parameters

	creds := Config.GetCredentials(config.CredentialRef.Name)
	identityStoreID := creds.GetString("identitystoreid")
	instanceARN := creds.GetString("instancearn")
	region := creds.GetString("region")

	group, ok := data["group"]
	if !ok {
		return AWSProviderParameters{}, fmt.Errorf("groupName not found in provider config")
	}

	username, ok := data["username"]
	if !ok {
		return AWSProviderParameters{}, fmt.Errorf("username not found in provider config")
	}

	return AWSProviderParameters{
		IdentityStoreID: identityStoreID,
		InstanceARN:     instanceARN,
		Region:          region,
		Group:           group,
		Username:        username,
	}, nil
}
