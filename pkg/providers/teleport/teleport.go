package teleport

import (
	"context"
	"errors"
	"fmt"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/gravitational/teleport/api/client"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
)

var Config = config.GetConfig()
var Tracer = otel.Tracer("pkg/providers/Teleport")

type TeleportProvider struct {
	Client     *client.Client
	Parameters TeleportProviderParameters
	Name       string `json:"name"`
}

type TeleportProviderParameters struct {
	Group           string `json:"group"`
	GroupDefinition string `json:"groupDefinition"`
	Username        string `json:"username"`
	CredentialsFile string `json:"credentialsFile"`
	Hostname        string `json:"hostname"`
}

func NewTeleportProvider(ctx context.Context, config models.ProviderConfig) (*TeleportProvider, error) {
	parameters, err := extractParameters(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider config: %w", err)
	}

	client, err := client.New(ctx, client.Config{
		Addrs: []string{
			parameters.Hostname,
		},
		Credentials: []client.Credentials{
			client.LoadIdentityFile(parameters.CredentialsFile),
		},
	})
	if err != nil {
		return nil, err
	}

	// defer client.Close()
	_, err = client.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &TeleportProvider{Parameters: parameters, Name: config.Name, Client: client}, nil
}

func (a *TeleportProvider) GrantAccess(ctx context.Context, request *models.AccessRequest) error {
	ctx, span := tracing.NewSpanWrapper(ctx, "providers.teleport.GrantAccess")
	defer span.End()

	parameters := a.Parameters

	if parameters.GroupDefinition != "" {
		err := a.upsertRole(parameters.Group, parameters.GroupDefinition)
		if err != nil {
			request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
			return fmt.Errorf("failed to create teleport role: %w", err)
		}
	}

	_, err := a.addRoleToUser(ctx, parameters.Username, parameters.Group)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to add user to group: %w", err)
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

func (a *TeleportProvider) RevokeAccess(ctx context.Context, request *models.AccessRequest) error {
	_, span := tracing.NewSpanWrapper(ctx, "providers.teleport.RevokeAccess")
	defer span.End()

	parameters := a.Parameters
	err := a.removeRoleFromUser(ctx, parameters.Username, parameters.Group)
	if err != nil {
		request.SetProviderStatusError(a.Name, parameters.Group, err.Error())
		return fmt.Errorf("failed to remove user from group: %w", err)
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

func (a *TeleportProvider) ListUsersWithAccess(ctx context.Context, roleRef models.AccessRoleRef) ([]string, error) {
	return []string{"user1", "user2"}, nil
}

func (a *TeleportProvider) IsAccessExpired(ctx context.Context, request *models.AccessRequest) (bool, error) {
	fmt.Printf("true")
	return true, nil
}

func extractParameters(config models.ProviderConfig) (TeleportProviderParameters, error) {

	data := config.Parameters

	creds := Config.GetCredentials(config.CredentialRef.Name)
	credentialsFile := creds.GetString("credentialsfile")
	hostname := creds.GetString("hostname")

	group, ok := data["group"]
	if !ok {
		return TeleportProviderParameters{}, errors.New("group not found in provider config")
	}

	username, ok := data["username"]
	if !ok {
		return TeleportProviderParameters{}, errors.New("username not found in provider config")
	}

	groupDefinition, ok := data["groupDefinition"]
	if !ok {
		groupDefinition = ""
	}

	return TeleportProviderParameters{
		Group:           group,
		GroupDefinition: groupDefinition,
		Username:        username,
		CredentialsFile: credentialsFile,
		Hostname:        hostname,
	}, nil
}
