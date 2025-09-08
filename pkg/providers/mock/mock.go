package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var Config = config.GetConfig()
var Tracer = otel.Tracer("pkg/providers/mock")

type MockProvider struct {
	Parameters MockProviderParameters
	Name       string `json:"name"`
}

type MockProviderParameters struct {
	Group    string `json:"group"`
	Username string `json:"username"`
}

func NewMockProvider(ctx context.Context, config models.ProviderConfig) (*MockProvider, error) {
	parameters, err := extractParameters(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider config: %w", err)
	}

	return &MockProvider{Parameters: parameters, Name: config.Name}, nil
}

func (a *MockProvider) GrantAccess(ctx context.Context, request *models.AccessRequest) error {
	_, span := tracing.NewSpanWrapper(ctx, "providers.mock.GrantAccess")
	span.SetAttributes(
		attribute.String("peer.service", "mock"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	parameters := a.Parameters

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

func (a *MockProvider) RevokeAccess(ctx context.Context, request *models.AccessRequest) error {
	_, span := tracing.NewSpanWrapper(ctx, "providers.mock.RevokeAccess")
	span.SetAttributes(
		attribute.String("peer.service", "mock"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	parameters := a.Parameters

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

func (a *MockProvider) ListUsersWithAccess(ctx context.Context, roleRef models.AccessRoleRef) ([]string, error) {
	return []string{"user1", "user2"}, nil
}

func (a *MockProvider) IsAccessExpired(ctx context.Context, request *models.AccessRequest) (bool, error) {
	fmt.Printf("true")
	return true, nil
}

func extractParameters(config models.ProviderConfig) (MockProviderParameters, error) {

	data := config.Parameters

	group, ok := data["group"]
	if !ok {
		return MockProviderParameters{}, errors.New("group not found in provider config")
	}

	username, ok := data["username"]
	if !ok {
		return MockProviderParameters{}, errors.New("username not found in provider config")
	}

	return MockProviderParameters{
		Group:    group,
		Username: username,
	}, nil
}
