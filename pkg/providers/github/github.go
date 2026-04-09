package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/providers/kinds"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"
	"golang.org/x/oauth2"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v74/github"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

var Config = config.GetConfig()
var Tracer = otel.Tracer("pkg/providers/github")

type GithubProvider struct {
	AppClient          *github.Client
	InstallationClient *github.Client
	PatClient          *github.Client
	Parameters         GithubProviderParameters
	Name               string `json:"name"`
}

type GithubProviderParameters struct {
	Org          string
	Role         string
	OrgRoles     []string
	Teams        map[string]string
	Repositories map[string]string
	RemoveUser   string
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
	appTr, err := ghinstallation.NewAppsTransportKeyFromFile(
		http.DefaultTransport,
		appID,
		keyPath,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create apps transport: %w", err)
	}

	appClient := github.NewClient(&http.Client{Transport: appTr})

	installations, _, err := appClient.Apps.ListInstallations(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list app installations for Github org: %s", parameters.Org)
	}

	for _, inst := range installations {
		if inst.GetAccount().GetLogin() == parameters.Org {
			installationID = inst.GetID()
			break
		}
	}

	if installationID == 0 {
		return nil, fmt.Errorf("could not find installation id for Github org: %s", parameters.Org)
	}

	// Installation client
	insTr := http.DefaultTransport
	insTr, err = ghinstallation.NewKeyFromFile(insTr, appID, installationID, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub App transport: %w", err)
	}

	tracedInsTr := otelhttp.NewTransport(insTr)
	installationClient := github.NewClient(&http.Client{Transport: tracedInsTr})

	// PAT Client
	pat := creds.GetString("pat")
	patTr := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, patTr)
	patClient := github.NewClient(tc)

	return &GithubProvider{
		AppClient:          appClient,
		InstallationClient: installationClient,
		PatClient:          patClient,
		Parameters:         parameters,
		Name:               config.Name,
	}, nil
}

func (p *GithubProvider) GrantAccess(ctx context.Context, request *models.AccessRequest) error {
	ctx, span := tracing.NewSpanWrapper(ctx, "providers.github.GrantAccess")
	defer span.End()

	params := p.Parameters
	username := request.GetProviderUsername(string(kinds.ProviderKindGithub))

	// Manage Org membership
	if params.Role != "" {
		err := p.addUserToOrg(ctx, params.Org, params.Role, username)
		if err != nil {
			request.SetProviderStatusError(p.Name, params.Role, err.Error())
			return fmt.Errorf("failed setting Github Org membership: %w", err)
		}
	}

	// Manage Org Roles
	if len(params.OrgRoles) > 0 {
		err := p.addOrgRolesToUser(ctx, params.Org, params.OrgRoles, username)
		if err != nil {
			return err
		}
	}

	// Manage Teams membership
	if len(params.Teams) > 0 {
		err := p.addUserToTeams(ctx, params.Org, params.Teams, username)
		if err != nil {
			request.SetProviderStatusError(p.Name, params.Role, err.Error())
			return fmt.Errorf("failed setting Github Teams membership %+v: %w", params.Teams, err)
		}
	}

	// Manage direct Repository access
	if len(params.Repositories) > 0 {
		err := p.addUserToRepos(ctx, params.Org, params.Repositories, username)
		if err != nil {
			request.SetProviderStatusError(p.Name, params.Role, err.Error())
			return fmt.Errorf("failed setting Direct Github Repository permissions %+v: %w", params.Teams, err)
		}

	}

	request.SetProviderStatusGranted(p.Name, params.Org, "")
	log.Info().
		Str("Provider", p.Name).
		Str("AccessRequest", request.Id).
		Str("Username", username).
		Str("Org", params.Org).
		Msg("User added to organization")
	return nil
}

func (p *GithubProvider) RevokeAccess(ctx context.Context, request *models.AccessRequest) error {
	ctx, span := tracing.NewSpanWrapper(ctx, "providers.github.RevokeAccess")
	defer span.End()

	params := p.Parameters
	username := request.GetProviderUsername(string(kinds.ProviderKindGithub))

	// Manage Org membership
	if params.RemoveUser == "true" {
		err := p.removeUserFromOrg(ctx, params.Org, username)
		if err != nil {
			request.SetProviderStatusError(p.Name, params.Org, err.Error())
			return fmt.Errorf("failed to remove user from org: %w", err)
		}
	}

	// Manage Org Roles
	if len(params.OrgRoles) > 0 {
		err := p.removeOrgRolesFromUser(ctx, params.Org, params.OrgRoles, username)
		if err != nil {
			return err
		}
	}

	// Manage Teams membership
	if len(params.Teams) > 0 {
		err := p.removeUserFromTeams(ctx, params.Org, params.Teams, username)
		if err != nil {
			request.SetProviderStatusError(p.Name, params.Role, err.Error())
			return fmt.Errorf("failed setting Github Teams membership %+v: %w", params.Teams, err)
		}
	}

	// Manage direct Repository access
	if len(params.Repositories) > 0 {
		err := p.removeUserFromRepos(ctx, params.Org, params.Repositories, username)
		if err != nil {
			request.SetProviderStatusError(p.Name, params.Role, err.Error())
			return fmt.Errorf("failed setting Direct Github Repository permissions %+v: %w", params.Teams, err)
		}

	}

	request.SetProviderStatusRevoked(p.Name, params.Org, "")
	log.Info().
		Str("Provider", p.Name).
		Str("AccessRequest", request.Id).
		Str("Username", username).
		Str("Org", params.Org).
		Msg("User access revoked")

	return nil
}

func (p *GithubProvider) ListUsersWithAccess(ctx context.Context, roleRef models.AccessRoleRef) ([]string, error) {
	ctx, span := tracing.NewSpanWrapper(ctx, "providers.github.ListUsersWithAccess")
	defer span.End()

	params := p.Parameters
	members, _, err := p.InstallationClient.Organizations.ListMembers(ctx, params.Org, nil)
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
