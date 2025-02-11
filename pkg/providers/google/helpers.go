package google

import (
	"context"
	"errors"

	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"

	"go.opentelemetry.io/otel/attribute"
)

func (g *GoogleProvider) addGroupMember(ctx context.Context, group string, username string) (err error) {
	ctx, span := tracing.NewSpanWrapper(ctx, "google.addGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "google"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	member := &admin.Member{
		Email: username,
		Role:  "MEMBER",
	}
	_, err = g.Service.Members.Insert(group, member).Context(ctx).Do()

	return err
}

func (g *GoogleProvider) removeGroupMember(ctx context.Context, group string, username string) (err error) {
	ctx, span := tracing.NewSpanWrapper(ctx, "google.removeGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "google"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	err = g.Service.Members.Delete(group, username).Context(ctx).Do()
	return err
}

// isGroupMember checks if a user is a member of the specified group
func (g *GoogleProvider) isGroupMember(ctx context.Context, Group, username string) (bool, error) {

	ctx, span := tracing.NewSpanWrapper(ctx, "google.isGroupMember")
	span.SetAttributes(
		attribute.String("peer.service", "google"),
		attribute.String("span.kind", "client"),
	)
	defer span.End()

	_, err := g.Service.Members.Get(Group, username).Context(ctx).Do()
	if err != nil {
		if googleapiError, ok := err.(*googleapi.Error); ok && googleapiError.Code == 404 {
			return false, nil // User is not a member
		}
		return false, err
	}
	return true, nil // User is a member
}

// extractParameters parses the provider config into GoogleProviderParameters
func extractParameters(config models.ProviderConfig) (GoogleProviderParameters, error) {

	data := config.Parameters

	creds := Config.GetCredentials(config.CredentialRef.Name)
	credentialsFile := creds.GetString("credentialsfile")

	group, ok := data["group"]
	if !ok {
		return GoogleProviderParameters{}, errors.New("group not found in provider config")
	}

	username, ok := data["username"]
	if !ok {
		return GoogleProviderParameters{}, errors.New("username not found in provider config")
	}

	return GoogleProviderParameters{
		Group:           group,
		Username:        username,
		CredentialsFile: credentialsFile,
	}, nil
}
