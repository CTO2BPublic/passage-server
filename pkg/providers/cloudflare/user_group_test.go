package cloudflare

import (
	"testing"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/iam"
	"github.com/stretchr/testify/assert"
)

func TestGroupPolicy_FromIAM(t *testing.T) {
	tests := []struct {
		name     string
		p        iam.UserGroupGetResponsePolicy
		expected *groupPolicy
	}{
		{
			name:     "empty",
			p:        iam.UserGroupGetResponsePolicy{},
			expected: &groupPolicy{},
		},
		{
			name: "with values",
			p: iam.UserGroupGetResponsePolicy{
				Access: "allow",
				PermissionGroups: []iam.UserGroupGetResponsePoliciesPermissionGroup{
					{ID: "pg1"},
				},
				ResourceGroups: []iam.UserGroupGetResponsePoliciesResourceGroup{
					{ID: "rg1"},
				},
			},
			expected: &groupPolicy{
				Access:           "allow",
				PermissionGroups: []permissionGroups{{ID: "pg1"}},
				ResourceGroups:   []resourceGroups{{ID: "rg1"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &groupPolicy{}
			g.fromIAM(tt.p)
			assert.Equal(t, tt.expected, g)
		})
	}
}

func TestGroupPolicy_ToNewPolicy(t *testing.T) {
	tests := []struct {
		name     string
		g        groupPolicy
		expected iam.UserGroupNewParamsPolicy
	}{
		{
			name: "empty",
			g:    groupPolicy{},
			expected: iam.UserGroupNewParamsPolicy{
				Access:           cloudflare.F(iam.UserGroupNewParamsPoliciesAccess("")),
				PermissionGroups: cloudflare.F([]iam.UserGroupNewParamsPoliciesPermissionGroup{}),
				ResourceGroups:   cloudflare.F([]iam.UserGroupNewParamsPoliciesResourceGroup{}),
			},
		},
		{
			name: "with values",
			g: groupPolicy{
				Access:           "allow",
				PermissionGroups: []permissionGroups{{ID: "pg1"}},
				ResourceGroups:   []resourceGroups{{ID: "rg1"}},
			},
			expected: iam.UserGroupNewParamsPolicy{
				Access: cloudflare.F(iam.UserGroupNewParamsPoliciesAccess("allow")),
				PermissionGroups: cloudflare.F([]iam.UserGroupNewParamsPoliciesPermissionGroup{
					{ID: cloudflare.F("pg1")},
				}),
				ResourceGroups: cloudflare.F([]iam.UserGroupNewParamsPoliciesResourceGroup{
					{ID: cloudflare.F("rg1")},
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.g.toNewPolicy()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupPolicy_ToUpdatePolicy(t *testing.T) {
	tests := []struct {
		name     string
		g        groupPolicy
		expected iam.UserGroupUpdateParamsPolicy
	}{
		{
			name: "empty",
			g:    groupPolicy{},
			expected: iam.UserGroupUpdateParamsPolicy{
				Access:           cloudflare.F(iam.UserGroupUpdateParamsPoliciesAccess("")),
				PermissionGroups: cloudflare.F([]iam.UserGroupUpdateParamsPoliciesPermissionGroup{}),
				ResourceGroups:   cloudflare.F([]iam.UserGroupUpdateParamsPoliciesResourceGroup{}),
			},
		},
		{
			name: "with values",
			g: groupPolicy{
				Access:           "allow",
				PermissionGroups: []permissionGroups{{ID: "pg1"}},
				ResourceGroups:   []resourceGroups{{ID: "rg1"}},
			},
			expected: iam.UserGroupUpdateParamsPolicy{
				Access: cloudflare.F(iam.UserGroupUpdateParamsPoliciesAccess("allow")),
				PermissionGroups: cloudflare.F([]iam.UserGroupUpdateParamsPoliciesPermissionGroup{
					{ID: cloudflare.F("pg1")},
				}),
				ResourceGroups: cloudflare.F([]iam.UserGroupUpdateParamsPoliciesResourceGroup{
					{ID: cloudflare.F("rg1")},
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.g.toUpdatePolicy()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupPolicy_Equal(t *testing.T) {
	basePolicy := groupPolicy{
		Access:           "allow",
		PermissionGroups: []permissionGroups{{ID: "pg1"}, {ID: "pg2"}},
		ResourceGroups:   []resourceGroups{{ID: "rg1"}, {ID: "rg2"}},
	}

	tests := []struct {
		name     string
		g        groupPolicy
		item     groupPolicy
		expected bool
	}{
		{
			name:     "equal policies",
			g:        basePolicy,
			item:     basePolicy,
			expected: true,
		},
		{
			name: "different access",
			g:    basePolicy,
			item: groupPolicy{
				Access:           "deny",
				PermissionGroups: basePolicy.PermissionGroups,
				ResourceGroups:   basePolicy.ResourceGroups,
			},
			expected: false,
		},
		{
			name: "different permission groups id",
			g:    basePolicy,
			item: groupPolicy{
				Access:           basePolicy.Access,
				PermissionGroups: []permissionGroups{{ID: "pg3"}, {ID: "pg2"}},
				ResourceGroups:   basePolicy.ResourceGroups,
			},
			expected: false,
		},
		{
			name: "different resource groups id",
			g:    basePolicy,
			item: groupPolicy{
				Access:           basePolicy.Access,
				PermissionGroups: basePolicy.PermissionGroups,
				ResourceGroups:   []resourceGroups{{ID: "rg3"}, {ID: "rg2"}},
			},
			expected: false,
		},
		{
			name: "different order permission groups",
			g:    basePolicy,
			item: groupPolicy{
				Access:           basePolicy.Access,
				PermissionGroups: []permissionGroups{{ID: "pg2"}, {ID: "pg1"}},
				ResourceGroups:   basePolicy.ResourceGroups,
			},
			expected: false,
		},
		{
			name: "different order resource groups",
			g:    basePolicy,
			item: groupPolicy{
				Access:           basePolicy.Access,
				PermissionGroups: basePolicy.PermissionGroups,
				ResourceGroups:   []resourceGroups{{ID: "rg2"}, {ID: "rg1"}},
			},
			expected: false,
		},
		{
			name: "different length permission groups",
			g:    basePolicy,
			item: groupPolicy{
				Access:           basePolicy.Access,
				PermissionGroups: []permissionGroups{{ID: "pg1"}},
				ResourceGroups:   basePolicy.ResourceGroups,
			},
			expected: false,
		},
		{
			name: "different length resource groups",
			g:    basePolicy,
			item: groupPolicy{
				Access:           basePolicy.Access,
				PermissionGroups: basePolicy.PermissionGroups,
				ResourceGroups:   []resourceGroups{{ID: "rg1"}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.g.Equal(tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}
