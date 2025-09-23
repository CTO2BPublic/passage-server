package github

import (
	"context"
	"testing"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/providers/kinds"

	"github.com/stretchr/testify/require"
)

func testProvider(t *testing.T) *GithubProvider {

	ctx := context.Background()
	c, err := NewGithubProvider(ctx, models.ProviderConfig{
		Name:     "Test",
		RunAsync: false,
		Provider: "github",
		CredentialRef: models.CredentialRef{
			Name: "github-test",
		},
		Parameters: map[string]string{
			"org":   "CTO2BPublic",
			"group": "admin",
		},
	})
	require.NoError(t, err)

	return c
}

func TestGrantRevoke(t *testing.T) {
	// t.Skip("skip until we have integration test environment setup")

	ctx := context.Background()
	c := testProvider(t)

	username := "test-user" // Replace with a test email
	providerType := string(kinds.ProviderKindGithub)

	err := c.GrantAccess(ctx, &models.AccessRequest{
		Status: models.AccessRequestStatus{
			ProviderUsernames: map[string]string{
				providerType: username,
			},
		},
	})
	require.NoError(t, err, "grant access should not fail")

	isMember, err := c.isOrgMember(ctx, c.Parameters.Org, username)
	require.NoError(t, err)
	require.True(t, isMember, "user should be a group member after grant")

	time.Sleep(time.Second * 10)
	err = c.RevokeAccess(ctx, &models.AccessRequest{
		Status: models.AccessRequestStatus{
			ProviderUsernames: map[string]string{
				providerType: username,
			},
		},
	})
	require.NoError(t, err)

	isMember, err = c.isOrgMember(ctx, c.Parameters.Org, username)
	require.NoError(t, err)
	require.False(t, isMember, "user should not be a group member after revoke")

}
