package controllers

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/dbdriver"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/providers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider is a mock implementation of the Provider interface
type mockProvider struct {
	revokeCalled bool
}

func (m *mockProvider) GrantAccess(ctx context.Context, request *models.AccessRequest) error {
	return nil
}

func (m *mockProvider) RevokeAccess(ctx context.Context, request *models.AccessRequest) error {
	m.revokeCalled = true
	return nil
}

func (m *mockProvider) ListUsersWithAccess(ctx context.Context, role models.AccessRoleRef) ([]string, error) {
	return nil, nil
}

func (m *mockProvider) IsAccessExpired(ctx context.Context, request *models.AccessRequest) (bool, error) {
	return false, nil
}

func setupHandlerTest(t *testing.T) (*gin.Context, *mockProvider, *dbdriver.Database) {

	// Setup test database
	dir, err := os.MkdirTemp("", "passage-test")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	Config.Db.Engine = "sqlite"
	Config.Db.Sqlite.Filename = dir + "/test.db"
	db := dbdriver.GetDriver()
	db.Connect()
	db.AutoMigrate()

	// Create a mock provider
	mockProv := &mockProvider{}
	// Override the NewProvider function to return the mock provider
	originalNewProvider := providers.NewProvider
	providers.SetNewProviderFactory(func(ctx context.Context, providerConfig models.ProviderConfig) (providers.Provider, error) {
		return mockProv, nil
	})
	t.Cleanup(func() {
		providers.SetNewProviderFactory(originalNewProvider)
	})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	return c, mockProv, db
}

func TestExpireAccessRequest_WithActiveRoles(t *testing.T) {
	c, mockProv, db := setupHandlerTest(t)

	accessRequests := []models.AccessRequest{
		{
			Id:      "request1",
			RoleRef: models.AccessRoleRef{Name: "test-role"},
			Status: models.AccessRequestStatus{
				RequestedBy: "test-user",
				Status:      models.AccessRequestApproved,
				ExpiresAt:   func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
				ProviderUsernames: map[string]string{
					"mock": "username",
				},
			},
		},
		{
			Id:      "request2",
			RoleRef: models.AccessRoleRef{Name: "test-role"},
			Status: models.AccessRequestStatus{
				RequestedBy: "test-user",
				Status:      models.AccessRequestApproved,
				ExpiresAt:   func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
				ProviderUsernames: map[string]string{
					"mock": "username",
				},
			},
		},
	}
	for _, ar := range accessRequests {
		err := db.InsertAccessRequest(context.Background(), ar)
		require.NoError(t, err)
	}

	// Create a new controller with the mock database
	controller := NewAccessRequestController()
	controller.Roles = []models.AccessRole{
		{
			Name: "test-role",
			Providers: []models.ProviderConfig{
				{
					Provider:   "mock",
					Parameters: make(map[string]string),
				},
			},
		},
	}

	// Create a test context
	c.Set("uid", "test-user")
	c.Set("utype", "token")
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "request1"}}

	// Expire the first request
	controller.Expire(c)

	// Assert that RevokeAccess was not called
	assert.False(t, mockProv.revokeCalled, "RevokeAccess should not have been called")

	// Expire the second request
	c.Params = gin.Params{gin.Param{Key: "ID", Value: "request2"}}
	controller.Expire(c)

	// Assert that RevokeAccess was called
	assert.True(t, mockProv.revokeCalled, "RevokeAccess should have been called")
}
