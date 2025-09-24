package github

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"
	"github.com/google/go-github/v74/github"
	"go.opentelemetry.io/otel/attribute"
	"gopkg.in/yaml.v2"
)

func (p *GithubProvider) isOrgMember(ctx context.Context, org, user string) (bool, error) {

	_, span := tracing.NewSpanWrapper(ctx, "github.isOrgMember")
	span.SetAttributes(
		attribute.String("peer.service", "github"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	membership, resp, err := p.InstallationClient.Organizations.GetOrgMembership(ctx, user, org)
	if resp != nil && resp.StatusCode == 404 {
		return false, nil
	}
	if err != nil {
		span.RecordError(err)
		return false, err
	}
	return membership != nil, nil
}

func (p *GithubProvider) addUserToOrg(ctx context.Context, org string, role string, username string) error {
	_, span := tracing.NewSpanWrapper(ctx, "github.addUserToOrg")
	span.SetAttributes(
		attribute.String("peer.service", "github"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	supportedRoles := []string{"member", "admin"}
	if !slices.Contains(supportedRoles, role) {
		err := fmt.Errorf("unsupported Github Org membership role: %s", role)
		span.RecordError(err)
		return err
	}

	_, _, err := p.InstallationClient.Organizations.EditOrgMembership(
		ctx,
		username,
		org,
		&github.Membership{Role: github.Ptr(role)},
	)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (p *GithubProvider) removeUserFromOrg(ctx context.Context, org string, username string) error {
	_, span := tracing.NewSpanWrapper(ctx, "github.removeUserFromOrg")
	span.SetAttributes(
		attribute.String("peer.service", "github"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	isMember, err := p.isOrgMember(ctx, org, username)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to check org membership: %w", err)
	}

	if isMember {
		_, err := p.InstallationClient.Organizations.RemoveOrgMembership(ctx, username, org)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	return nil
}

func (p *GithubProvider) addUserToTeams(ctx context.Context, org string, teams map[string]string, username string) error {
	_, span := tracing.NewSpanWrapper(ctx, "github.addUserToTeams")
	span.SetAttributes(
		attribute.String("peer.service", "github"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	for team, role := range teams {

		supportedRoles := []string{"member", "maintainer"}
		if !slices.Contains(supportedRoles, role) {
			err := fmt.Errorf("unsupported Github Team membership role: %s", role)
			span.RecordError(err)
			return err
		}

		_, _, err := p.InstallationClient.Teams.AddTeamMembershipBySlug(
			ctx,
			org,
			team,
			username,
			&github.TeamAddTeamMembershipOptions{
				Role: role,
			},
		)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	return nil
}

func (p *GithubProvider) removeUserFromTeams(ctx context.Context, org string, teams map[string]string, username string) error {
	_, span := tracing.NewSpanWrapper(ctx, "removeUserFromTeams.removeUserFromTeams")
	span.SetAttributes(
		attribute.String("peer.service", "github"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	for team := range teams {
		_, err := p.InstallationClient.Teams.RemoveTeamMembershipBySlug(
			ctx,
			org,
			team,
			username,
		)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	return nil
}

func (p *GithubProvider) addOrgRolesToUser(ctx context.Context, org string, orgRoles []string, username string) error {
	_, span := tracing.NewSpanWrapper(ctx, "github.addOrgRolesToUser")
	span.SetAttributes(
		attribute.String("peer.service", "github"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	// Fetch existing custom roles
	existingRoles, _, err := p.InstallationClient.Organizations.ListRoles(ctx, org)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Assign roles to user
	for _, role := range orgRoles {

		var roleId int64
		found := false

		for _, r := range existingRoles.CustomRepoRoles {
			if r.GetName() == role {
				roleId = r.GetID()
				found = true
				break
			}
		}

		if !found {
			var available []string
			for _, r := range existingRoles.CustomRepoRoles {
				available = append(available, r.GetName())
			}
			err = fmt.Errorf("organization role %q not found. Available roles: %s", role, strings.Join(available, ", "))
			span.RecordError(err)
			return err
		}

		_, err := p.InstallationClient.Organizations.AssignOrgRoleToUser(
			ctx,
			org,
			username,
			roleId,
		)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	return nil
}

func (p *GithubProvider) removeOrgRolesFromUser(ctx context.Context, org string, orgRoles []string, username string) error {
	_, span := tracing.NewSpanWrapper(ctx, "github.removeOrgRolesFromUser")
	span.SetAttributes(
		attribute.String("peer.service", "github"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	// Fetch existing custom roles
	existingRoles, _, err := p.InstallationClient.Organizations.ListRoles(ctx, org)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Remove roles from user
	for _, role := range orgRoles {

		var roleId int64

		for _, r := range existingRoles.CustomRepoRoles {
			if r.GetName() == role {
				roleId = r.GetID()
				break
			}
		}

		_, err := p.InstallationClient.Organizations.RemoveOrgRoleFromUser(
			ctx,
			org,
			username,
			roleId,
		)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	return nil
}

func (p *GithubProvider) addUserToRepos(ctx context.Context, org string, repos map[string]string, username string) error {
	_, span := tracing.NewSpanWrapper(ctx, "github.addUserToRepos")
	span.SetAttributes(
		attribute.String("peer.service", "github"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	for repo, role := range repos {

		_, _, err := p.InstallationClient.Repositories.AddCollaborator(
			ctx,
			org,
			repo,
			username,
			&github.RepositoryAddCollaboratorOptions{
				Permission: role,
			},
		)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	return nil
}

func (p *GithubProvider) removeUserFromRepos(ctx context.Context, org string, repos map[string]string, username string) error {
	_, span := tracing.NewSpanWrapper(ctx, "github.removeUserFromRepos")
	span.SetAttributes(
		attribute.String("peer.service", "github"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	for repo := range repos {

		_, err := p.InstallationClient.Repositories.RemoveCollaborator(
			ctx,
			org,
			repo,
			username,
		)
		if err != nil {
			span.RecordError(err)
			return err
		}
	}

	return nil
}

func extractParameters(cfg models.ProviderConfig) (GithubProviderParameters, error) {
	data := cfg.Parameters

	org, ok := data["org"]
	if !ok {
		return GithubProviderParameters{}, errors.New("org not found in provider config")
	}

	role, ok := data["role"]
	if !ok {
		role = ""
	}

	orgRolesList := []string{}
	orgRoles, ok := data["orgRoles"]
	if ok {
		err := yaml.Unmarshal([]byte(orgRoles), &orgRolesList)
		if err != nil {
			return GithubProviderParameters{}, fmt.Errorf("failed to unmarshal access roles: %w", err)
		}
	}

	teamsMap := map[string]string{}
	teams, ok := data["teams"]
	if ok {
		err := yaml.Unmarshal([]byte(teams), &teamsMap)
		if err != nil {
			return GithubProviderParameters{}, fmt.Errorf("failed to unmarshal teams: %w", err)
		}
	}

	repoMap := map[string]string{}
	repositories, ok := data["repositories"]
	if ok {
		err := yaml.Unmarshal([]byte(repositories), &repoMap)
		if err != nil {
			return GithubProviderParameters{}, fmt.Errorf("failed to unmarshal repositories: %w", err)
		}
	}

	removeUser, ok := data["removeUser"]
	if !ok {
		removeUser = "false"
	}

	return GithubProviderParameters{
		Org:          org,
		Role:         role,
		OrgRoles:     orgRolesList,
		Teams:        teamsMap,
		Repositories: repoMap,
		RemoveUser:   removeUser,
	}, nil
}

func parseInt64(v string) (int64, error) {
	return strconv.ParseInt(v, 10, 64)
}
