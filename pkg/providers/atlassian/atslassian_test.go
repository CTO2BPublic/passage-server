package atlassian

import (
	"context"
	"testing"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testProvider(t *testing.T) *AtlassianProvider {
	cfg := config.GetConfig()
	cfg.Creds = map[string]models.Credential{
		"test": {
			Name: "test",
			Data: map[string]string{
				"credentialsfile": "creds/jira-token",     // Replace with real path to token file
				"email":           "admin-user@email.com", // Replace with real admin email
			},
		},
	}

	ctx := context.Background()
	c, err := NewAtlassianProvider(ctx, models.ProviderConfig{
		Name:     "Test",
		RunAsync: false,
		Provider: "atlassian",
		CredentialRef: models.CredentialRef{
			Name: "test",
		},
		Parameters: map[string]string{
			"group":   "jira-users-todelete-test",   // This group must already exist in your Jira instance
			"siteURL": "https://test.atlassian.net", // Replace with your Jira site URL
		},
	})
	require.NoError(t, err)

	return c
}

func TestGrantRevoke(t *testing.T) {
	t.Skip("skip until we have integration test environment setup")

	ctx := context.Background()
	c := testProvider(t)

	email := "newuser@email.com"
	group := "test-group"

	_, err := c.getUser(ctx, email)
	require.NoError(t, err, "test user should exist in the system already")

	isMember, err := c.isGroupMember(ctx, group, email)
	require.NoError(t, err)
	assert.False(t, isMember, "user should not be a member of the group yet")

	err = c.addGroupMember(ctx, group, email)
	assert.NoError(t, err, "adding user to group should not error")

	isMember, err = c.isGroupMember(ctx, group, email)
	require.NoError(t, err)
	assert.True(t, isMember, "user should be a member of the group now")

	err = c.removeGroupMember(ctx, group, email)
	assert.NoError(t, err, "removing user from group should not error")

	isMember, err = c.isGroupMember(ctx, group, email)
	require.NoError(t, err)
	assert.False(t, isMember, "user should not be a member of the group anymore")
}

func TestGroupMembers(t *testing.T) {
	t.Skip("skip until we have integration test environment setup")

	ctx := context.Background()
	c := testProvider(t)

	group := "org-admins"
	// the email used by the client should be an admin
	email, _ := c.client.Auth.GetBasicAuth()

	members, err := c.getGroupMembers(ctx, group)
	require.NoError(t, err)
	assert.Greater(t, len(members), 0, "there should be at least one member in the admin group")

	found := false
	for _, m := range members {
		if m.EmailAddress == email {
			found = true
		}
	}
	assert.True(t, found, "the authenticated user should be in the admin group")
}
