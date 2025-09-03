package providers

import (
	"context"
	"fmt"

	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/providers/aws"
	"github.com/CTO2BPublic/passage-server/pkg/providers/cloudflare"
	"github.com/CTO2BPublic/passage-server/pkg/providers/gitlab"
	"github.com/CTO2BPublic/passage-server/pkg/providers/google"
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
	case "mock":
		return mockp.NewMockProvider(ctx, providerConfig)
	case "aws":
		return aws.NewAWSProvider(ctx, providerConfig)
	case "gitlab":
		return gitlab.NewGitlabProvider(ctx, providerConfig)
	case "google":
		return google.NewGoogleProvider(ctx, providerConfig)
	case "teleport":
		return teleport.NewTeleportProvider(ctx, providerConfig)
	case "cloudflare":
		return cloudflare.NewCloudflareProvider(ctx, providerConfig)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerConfig.Provider)
	}
}

func NewProviderUsernames() models.ProviderUsernames {
	return models.ProviderUsernames{
		ProviderUsernames: map[string]string{
			"gitlab":   "",
			"teleport": "",
			"google":   "",
			"aws":      "",
			"azure":    "",
		},
	}
}
