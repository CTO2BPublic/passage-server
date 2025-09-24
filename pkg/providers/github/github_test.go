package github

import (
	"context"
	"testing"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/providers/kinds"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"
	"github.com/rs/zerolog/log"

	"github.com/stretchr/testify/require"
)

func testProvider(t *testing.T, params map[string]string) (*GithubProvider, error) {

	err := config.InitConfig("../../../configs")
	if _, err := tracing.NewTracer(); err != nil {
		log.Fatal().Err(err).Msg("Failed to read config")
	}

	Config := config.GetConfig()

	if Config.Tracing.Enabled {
		if _, err := tracing.NewTracer(); err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize tracer")
		}
	}

	ctx := context.Background()
	c, err := NewGithubProvider(ctx, models.ProviderConfig{
		Name:     "Test",
		RunAsync: false,
		Provider: "github",
		CredentialRef: models.CredentialRef{

			Name: "test-github",
		},
		Parameters: params,
	})

	return c, err
}

func mustProvider(t *testing.T, params map[string]string) *GithubProvider {
	p, err := testProvider(t, params)
	require.NoError(t, err)
	return p
}

type grantAccessTestCase struct {
	name        string
	params      map[string]string
	wantError   bool
	errContain  string
	checkRevoke bool // whether to call RevokeAccess after Grant
}

func TestGrantAccessVariousConfigs(t *testing.T) {
	ctx := context.Background()
	username := "cto2bserviceaccount"
	providerType := string(kinds.ProviderKindGithub)

	cases := []grantAccessTestCase{
		{
			name: "valid config",
			params: map[string]string{
				"org":          "CTO2BPublic",
				"role":         "admin",
				"orgRoles":     `["all_repo_read"]`,
				"teams":        `{"cto2bprimary":"member"}`,
				"repositories": `{"office-supplies-tracker":"admin"}`,
				"removeUser":   "true",
			},
			wantError:   false,
			checkRevoke: true,
		},
		{
			name: "external collaborator",
			params: map[string]string{
				"org":          "CTO2BPublic",
				"repositories": `{"office-supplies-tracker":"push"}`,
			},
			wantError:   false,
			checkRevoke: true,
		},
		{
			name:       "invalid org role",
			params:     map[string]string{"org": "CTO2BPublic", "orgRoles": `["Non existant org role"]`},
			wantError:  true,
			errContain: "not found",
		},
		{
			name:       "invalid org membership",
			params:     map[string]string{"org": "CTO2BPublic", "role": "NonExistantTestRole"},
			wantError:  true,
			errContain: "unsupported",
		},
		{
			name:       "invalid org",
			params:     map[string]string{"org": "ctobpublicnonexistant"},
			wantError:  true,
			errContain: "could not find installation id",
		},
		{
			name: "invalid team permissions",
			params: map[string]string{
				"org":   "CTO2BPublic",
				"teams": `{"cto2bprimary":"membertest"}`,
			},
			wantError:   true,
			errContain:  "unsupported Github Team membership role",
			checkRevoke: true,
		},
		{
			name: "invalid repository permissions",
			params: map[string]string{
				"org":          "CTO2BPublic",
				"repositories": `{"office-supplies-tracker":"pushtest"}`,
			},
			wantError:  true,
			errContain: "is not a valid permission",
		},
		{
			name: "invalid team",
			params: map[string]string{
				"org":   "CTO2BPublic",
				"teams": `{"cto2bprimarytest":"member"}`,
			},
			wantError:  true,
			errContain: "404 Not Found",
		},
		{
			name: "invalid repository",
			params: map[string]string{
				"org":          "CTO2BPublic",
				"repositories": `{"office-supplies-tracker-test":"pushtest"}`,
			},
			wantError:  true,
			errContain: "404 Not Found",
		},
		// add more cases as needed
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			c, err := testProvider(t, tt.params)

			// special handling for invalid org (provider creation fails)
			if tt.wantError && tt.errContain == "could not find installation id" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContain)
				return
			}
			require.NoError(t, err)

			req := &models.AccessRequest{
				Status: models.AccessRequestStatus{
					ProviderUsernames: map[string]string{providerType: username},
				},
			}

			err = c.GrantAccess(ctx, req)
			if tt.wantError {
				require.Error(t, err)
				if tt.errContain != "" {
					require.Contains(t, err.Error(), tt.errContain)
				}
			} else {
				require.NoError(t, err)
			}

			if !tt.wantError && tt.checkRevoke {
				err = c.RevokeAccess(ctx, req)
				require.NoError(t, err)
			}
		})
	}
}

func TestIsAccessExpired(t *testing.T) {
	ctx := context.Background()
	params := map[string]string{
		"org":          "CTO2BPublic",
		"repositories": `{"office-supplies-tracker-test":"pushtest"}`,
	}
	c := mustProvider(t, params)

	now := time.Now()

	tests := []struct {
		name        string
		ttl         string
		createdAt   time.Time
		wantExpired bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "no TTL specified",
			ttl:         "",
			createdAt:   now,
			wantExpired: false,
			wantErr:     true,
			errContains: "TTL not specified",
		},
		{
			name:        "invalid TTL format",
			ttl:         "notaduration",
			createdAt:   now,
			wantExpired: false,
			wantErr:     true,
			errContains: "invalid TTL format",
		},
		{
			name:        "not expired",
			ttl:         "1h",
			createdAt:   now,
			wantExpired: false,
			wantErr:     false,
		},
		{
			name:        "expired",
			ttl:         "1s",
			createdAt:   now.Add(-2 * time.Second),
			wantExpired: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &models.AccessRequest{
				CreatedAt: tt.createdAt,
				Details: models.AccessRequestDetails{
					TTL: tt.ttl,
				},
			}

			expired, err := c.IsAccessExpired(ctx, req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantExpired, expired)
			}
		})
	}
}
