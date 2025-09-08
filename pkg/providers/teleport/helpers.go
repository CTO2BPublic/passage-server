package teleport

import (
	"context"
	"slices"
	"strings"

	"github.com/CTO2BPublic/passage-server/pkg/tracing"
	"github.com/gravitational/teleport/api/types"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v2"

	"go.opentelemetry.io/otel/attribute"
)

func (a *TeleportProvider) addRoleToUser(ctx context.Context, Username string, RoleName string) (Created bool, Error error) {
	ctx, span := tracing.NewSpanWrapper(ctx, "teleport.addRoleToUser")
	span.SetAttributes(
		attribute.String("peer.service", "teleport"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	user, err := a.Client.GetUser(ctx, Username, false)
	if err != nil {
		return false, err
	}

	if user.GetName() == Username {

		roles := a.parseRoles(ctx, RoleName)

		for _, role := range roles {
			user.AddRole(role)
		}

		_, err := a.Client.UpdateUser(ctx, user)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (a *TeleportProvider) upsertRole(RoleName string, roleDefinition string) error {
	ctx := context.Background()
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	var holder map[interface{}]interface{}
	if err := yaml.Unmarshal([]byte(roleDefinition), &holder); err != nil {
		return err
	}
	bytes, err := json.Marshal(holder)
	if err != nil {
		return err
	}

	var role types.RoleV6
	err = json.Unmarshal(bytes, &role)
	if err != nil {
		return err
	}

	role.SetMetadata(types.Metadata{
		Name:        RoleName,
		Description: "Passage created role",
	})

	if a.Client != nil {

		_, err = a.Client.UpsertRole(ctx, &role)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

func (a *TeleportProvider) removeRoleFromUser(ctx context.Context, Username string, RoleName string) error {
	ctx, span := tracing.NewSpanWrapper(ctx, "teleport.removeRoleFromUser")
	span.SetAttributes(
		attribute.String("peer.service", "teleport"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	user, err := a.Client.GetUser(ctx, Username, false)
	if err != nil {
		return err
	}

	filteredRoles := []string{}

	currentRoles := user.GetRoles()
	rolesToRemove := a.parseRoles(ctx, RoleName)

	for _, role := range currentRoles {
		if !slices.Contains(rolesToRemove, role) {
			filteredRoles = append(filteredRoles, role)
		}
	}

	user.SetRoles(filteredRoles)

	_, err = a.Client.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (a *TeleportProvider) parseRoles(ctx context.Context, group string) []string {

	roles := strings.Split(group, ",")

	if len(roles) > 0 {
		return roles
	}

	return []string{"group"}
}
