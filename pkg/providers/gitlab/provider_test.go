package gitlab_test

import (
	"context"
	"testing"

	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/providers/gitlab"
)

func TestGitlabProvider(t *testing.T) {
	ctx := context.Background()
	p, err := gitlab.NewGitlabProvider(ctx, models.ProviderConfig{})
	if err != nil {
		t.Error("Failed to init provider")
	}

	err = p.GrantAccess(context.Background(), &models.AccessRequest{})
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}
}
