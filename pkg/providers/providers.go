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

type ProviderKind string

const (
	ProviderKindMock       ProviderKind = "mock"
	ProviderKindGitlab     ProviderKind = "gitlab"
	ProviderKindGoogle     ProviderKind = "google"
	ProviderKindTeleport   ProviderKind = "teleport"
	ProviderKindAWS        ProviderKind = "aws"
	ProviderKindCloudflare ProviderKind = "cloudflare"
)

var allProviderKinds = []ProviderKind{
	ProviderKindMock,
	ProviderKindGitlab,
	ProviderKindGoogle,
	ProviderKindTeleport,
	ProviderKindAWS,
	ProviderKindCloudflare,
}

func NewProvider(ctx context.Context, providerConfig models.ProviderConfig) (Provider, error) {

	switch providerConfig.Provider {
	case string(ProviderKindMock):
		return mockp.NewMockProvider(ctx, providerConfig)
	case string(ProviderKindAWS):
		return aws.NewAWSProvider(ctx, providerConfig)
	case string(ProviderKindGitlab):
		return gitlab.NewGitlabProvider(ctx, providerConfig)
	case string(ProviderKindGoogle):
		return google.NewGoogleProvider(ctx, providerConfig)
	case string(ProviderKindTeleport):
		return teleport.NewTeleportProvider(ctx, providerConfig)
	case string(ProviderKindCloudflare):
		return cloudflare.NewCloudflareProvider(ctx, providerConfig)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerConfig.Provider)
	}
}

func NewProviderUsernames() models.ProviderUsernames {
	p := models.ProviderUsernames{
		ProviderUsernames: make(map[string]string, len(allProviderKinds)),
	}
	for _, kind := range allProviderKinds {
		// Initialize with empty string
		// This ensures that all providers are present in the map
		// even if they are not used
		p.ProviderUsernames[string(kind)] = ""
	}
	return p
}
