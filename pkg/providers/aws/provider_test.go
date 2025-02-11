package aws_test

import (
	"context"
	"testing"

	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/providers/aws"
)

func TestGoogleProvider(t *testing.T) {
	ctx := context.Background()
	p, err := aws.NewAWSProvider(ctx, models.ProviderConfig{})
	if err != nil {
		t.Error("Failed to init provider")
	}

	err = p.GrantAccess(context.Background(), &models.AccessRequest{})
	if err != nil {
		t.Errorf("Failed to add user: %v", err)
	}
}
