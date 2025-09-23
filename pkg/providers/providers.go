package providers

import (
	"context"
	"fmt"

	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/providers/aws"
	"github.com/CTO2BPublic/passage-server/pkg/providers/cloudflare"
	"github.com/CTO2BPublic/passage-server/pkg/providers/github"
	"github.com/CTO2BPublic/passage-server/pkg/providers/gitlab"
	"github.com/CTO2BPublic/passage-server/pkg/providers/google"
	"github.com/CTO2BPublic/passage-server/pkg/providers/kinds"
	mockp "github.com/CTO2BPublic/passage-server/pkg/providers/mock"
	"github.com/CTO2BPublic/passage-server/pkg/providers/teleport"
)

type Provider interface {
	GrantAccess(ctx context.Context, request *models.AccessRequest) error
	RevokeAccess(ctx context.Context, request *models.AccessRequest) error
	ListUsersWithAccess(ctx context.Context, role models.AccessRoleRef) ([]string, error)
	IsAccessExpired(ctx context.Context, request *models.AccessRequest) (bool, error)
}

func NewProvider(ctx context.Context, providerConfig models.ProviderConfig) (Provider, error) {

	switch providerConfig.Provider {
	case string(kinds.ProviderKindMock):
		return mockp.NewMockProvider(ctx, providerConfig)
	case string(kinds.ProviderKindAWS):
		return aws.NewAWSProvider(ctx, providerConfig)
	case string(kinds.ProviderKindGitlab):
		return gitlab.NewGitlabProvider(ctx, providerConfig)
	case string(kinds.ProviderKindGoogle):
		return google.NewGoogleProvider(ctx, providerConfig)
	case string(kinds.ProviderKindTeleport):
		return teleport.NewTeleportProvider(ctx, providerConfig)
	case string(kinds.ProviderKindCloudflare):
		return cloudflare.NewCloudflareProvider(ctx, providerConfig)
	case string(kinds.ProviderKindGithub):
		return github.NewGithubProvider(ctx, providerConfig)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerConfig.Provider)
	}
}

func NewProviderUsernames() models.ProviderUsernames {
	p := models.ProviderUsernames{
		ProviderUsernames: make(map[string]string, len(kinds.AllProviderKinds)),
	}
	for _, kind := range kinds.AllProviderKinds {
		// Initialize with empty string
		// This ensures that all providers are present in the map
		// even if they are not used
		p.ProviderUsernames[string(kind)] = ""
	}
	return p
}
