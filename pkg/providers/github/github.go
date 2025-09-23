package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v74/github"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
)

var Config = config.GetConfig()
var Tracer = otel.Tracer("pkg/providers/github")

type GithubProvider struct {
	Client     *github.Client
	Parameters GithubProviderParameters
	Name       string `json:"name"`
}

type GithubProviderParameters struct {
	Org      string `json:"org"`
	Username string `json:"username"`
	Group    string `json:"group"` // admin, member
}

func NewGithubProvider(ctx context.Context, config models.ProviderConfig) (*GithubProvider, error) {
	parameters, err := extractParameters(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider config: %w", err)
	}

	creds := Config.GetCredentials(config.CredentialRef.Name)

	appID, err := parseInt64(creds.GetString("appid"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse appid: %w", err)
	}

	keyPath := creds.GetString("privatekeypath")

	var installationID int64

	// App client
	tr, _ := ghinstallation.NewAppsTransportKeyFromFile(
		http.DefaultTransport,
		appID,
		keyPath,
	)

	appClient := github.NewClient(&http.Client{Transport: tr})

	installations, _, _ := appClient.Apps.ListInstallations(ctx, nil)
	for _, inst := range installations {
		if inst.GetAccount().GetLogin() == parameters.Org {
			installationID = inst.GetID()
			log.Debug().Msgf("ID: %d Account: %s", inst.GetID(), inst.GetAccount().GetLogin())
			continue
		}
	}

	if installationID == 0 {
		return nil, fmt.Errorf("could not find installation id for Github org: %s", parameters.Org)
	}

	// Installation client
	itr := http.DefaultTransport
	itr, err = ghinstallation.NewKeyFromFile(itr, appID, installationID, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub App transport: %w", err)
	}

	client := github.NewClient(&http.Client{Transport: itr})

	return &GithubProvider{
		Client:     client,
		Parameters: parameters,
		Name:       config.Name,
	}, nil
}

func (p *GithubProvider) GrantAccess(ctx context.Context, request *models.AccessRequest) error {
	ctx, span := tracing.NewSpanWrapper(ctx, "providers.github.GrantAccess")
	defer span.End()

	params := p.Parameters

	_, _, err := p.Client.Organizations.EditOrgMembership(ctx, params.Username, params.Org, &github.Membership{Role: github.Ptr(params.Group)})
	if err != nil {
		request.SetProviderStatusError(p.Name, params.Org, err.Error())
		return fmt.Errorf("failed to add user to org: %w", err)
	}

	request.SetProviderStatusGranted(p.Name, params.Org, "")
	log.Info().
		Str("Provider", p.Name).
		Str("AccessRequest", request.Id).
		Str("Username", params.Username).
		Str("Org", params.Org).
		Msg("User added to organization")
	return nil
}

func (p *GithubProvider) RevokeAccess(ctx context.Context, request *models.AccessRequest) error {
	ctx, span := tracing.NewSpanWrapper(ctx, "providers.github.RevokeAccess")
	defer span.End()

	params := p.Parameters

	isMember, err := p.isOrgMember(ctx, params.Org, params.Username)
	if err != nil {
		request.SetProviderStatusError(p.Name, params.Org, err.Error())
		return fmt.Errorf("failed to check org membership: %w", err)
	}

	if !isMember {
		request.SetProviderStatusRevoked(p.Name, params.Org, "already removed")
		log.Info().
			Str("Provider", p.Name).
			Str("AccessRequest", request.Id).
			Str("Username", params.Username).
			Str("Org", params.Org).
			Msg("User already not in organization")
		return nil
	}

	_, err = p.Client.Organizations.RemoveOrgMembership(ctx, params.Username, params.Org)
	if err != nil {
		request.SetProviderStatusError(p.Name, params.Org, err.Error())
		return fmt.Errorf("failed to remove user from org: %w", err)
	}

	request.SetProviderStatusRevoked(p.Name, params.Org, "")
	log.Info().
		Str("Provider", p.Name).
		Str("AccessRequest", request.Id).
		Str("Username", params.Username).
		Str("Org", params.Org).
		Msg("User removed from organization")
	return nil
}

func (p *GithubProvider) ListUsersWithAccess(ctx context.Context, roleRef models.AccessRoleRef) ([]string, error) {
	ctx, span := tracing.NewSpanWrapper(ctx, "providers.github.ListUsersWithAccess")
	defer span.End()

	params := p.Parameters
	members, _, err := p.Client.Organizations.ListMembers(ctx, params.Org, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list org members: %w", err)
	}

	usernames := []string{}
	for _, m := range members {
		if m.Login != nil {
			usernames = append(usernames, *m.Login)
		}
	}
	return usernames, nil
}

func (p *GithubProvider) IsAccessExpired(ctx context.Context, request *models.AccessRequest) (bool, error) {
	ttl := request.Details.TTL
	if ttl == "" {
		return false, errors.New("TTL not specified")
	}
	expiry, err := time.ParseDuration(ttl)
	if err != nil {
		return false, fmt.Errorf("invalid TTL format: %w", err)
	}
	expirationTime := request.CreatedAt.Add(expiry)
	return time.Now().After(expirationTime), nil
}
