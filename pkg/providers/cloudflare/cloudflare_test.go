package cloudflare

import (
	"context"
	"testing"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testProvider(t *testing.T) *CloudflareProvider {
	cfg := config.GetConfig()
	cfg.Creds = map[string]models.Credential{
		"test": {
			Name: "test",
			Data: map[string]string{
				"credentialsfile": "./creds/cloudflare-api-token",
			},
		},
	}

	ctx := context.Background()
	c, err := NewCloudflareProvider(ctx, models.ProviderConfig{
		Name:     "Test",
		RunAsync: false,
		Provider: "cloudflare",
		CredentialRef: models.CredentialRef{
			Name: "test",
		},
		Parameters: map[string]string{
			"accountID": "", // Replace with your Cloudflare Account ID
			"group":     "", // Replace with the name of the test group
		},
	})
	require.NoError(t, err)

	return c
}

func TestGrantRevoke(t *testing.T) {
	t.Skip("skip until we have integration test environment setup")

	ctx := context.Background()
	c := testProvider(t)

	email := "newuser@email.com" // Replace with a test email

	err := c.GrantAccess(ctx, &models.AccessRequest{
		Status: models.AccessRequestStatus{
			ProviderUsernames: map[string]string{
				providerType: email,
			},
		},
	})
	require.NoError(t, err, "grant access should not fail")

	isMember, err := c.isGroupMember(ctx, c.groupID, email)
	require.NoError(t, err)
	require.True(t, isMember, "user should be a group member after grant")

	err = c.RevokeAccess(ctx, &models.AccessRequest{
		Status: models.AccessRequestStatus{
			ProviderUsernames: map[string]string{
				providerType: email,
			},
		},
	})
	require.NoError(t, err)

	isMember, err = c.isGroupMember(ctx, c.groupID, email)
	require.NoError(t, err)
	require.False(t, isMember, "user should not be a group member after revoke")

}

func TestListUsersWithAccess(t *testing.T) {
	t.Skip("skip until we have integration test environment setup")

	ctx := context.Background()
	c := testProvider(t)
	email := "newuser@email.com" // Replace with a test email

	members, err := c.ListUsersWithAccess(ctx, models.AccessRoleRef{
		Name: c.groupName,
	})
	require.NoError(t, err)
	assert.Empty(t, members)

	err = c.addGroupMember(ctx, c.groupID, email)
	require.NoError(t, err)

	members, err = c.ListUsersWithAccess(ctx, models.AccessRoleRef{
		Name: c.groupName,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, members)

	// get back to original state
	err = c.removeGroupMember(ctx, c.groupID, email)
	require.NoError(t, err)
}
