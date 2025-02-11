package aws_test

import (
	"context"
	"testing"

	"github.com/CTO2BPublic/passage-server/pkg/models"
	mockp "github.com/CTO2BPublic/passage-server/pkg/providers/mock"
)

func TestAWSProvider(t *testing.T) {
	ctx := context.Background()
	p, err := mockp.NewMockProvider(ctx, models.ProviderConfig{})
	if err != nil {
		t.Error("Failed to init provider")
	}

	err = p.GrantAccess(context.Background(), &models.AccessRequest{})
	if err != nil {
		t.Errorf("Failed to add user: %v", err)
	}
}
