package github

import (
	"context"
	"errors"
	"strconv"

	"github.com/CTO2BPublic/passage-server/pkg/models"
)

func (p *GithubProvider) isOrgMember(ctx context.Context, org, user string) (bool, error) {
	membership, resp, err := p.InstallationClient.Organizations.GetOrgMembership(ctx, user, org)
	if resp != nil && resp.StatusCode == 404 {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return membership != nil, nil
}

func extractParameters(cfg models.ProviderConfig) (GithubProviderParameters, error) {
	data := cfg.Parameters

	org, ok := data["org"]
	if !ok {
		return GithubProviderParameters{}, errors.New("org not found in provider config")
	}
	group, ok := data["group"]
	if !ok {
		return GithubProviderParameters{}, errors.New("group not found in provider config")
	}

	return GithubProviderParameters{
		Org:   org,
		Group: group,
	}, nil
}

func parseInt64(v string) (int64, error) {
	return strconv.ParseInt(v, 10, 64)
}
