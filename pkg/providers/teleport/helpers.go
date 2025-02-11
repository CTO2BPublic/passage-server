package teleport

import (
	"context"

	"github.com/CTO2BPublic/passage-server/pkg/tracing"

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

		user.AddRole(RoleName)

		_, err := a.Client.UpdateUser(ctx, user)
		if err != nil {
			return false, err
		}
	}
	return true, nil
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
	roles := user.GetRoles()

	for _, role := range roles {
		if role != RoleName {
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
