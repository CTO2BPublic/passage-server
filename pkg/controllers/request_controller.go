package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/CTO2BPublic/passage-server/pkg/errors"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/providers"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/rs/zerolog/log"
)

type AccessRequestController struct {
	Providers     map[string]models.ProviderConfig
	Creds         map[string]models.Credential
	Roles         []models.AccessRole
	ApprovalRules []models.ApprovalRule
}

func NewAccessRequestController() *AccessRequestController {

	controller := AccessRequestController{}

	// Load roles
	controller.Roles = Config.Roles
	controller.Creds = Config.Creds
	controller.ApprovalRules = Config.ApprovalRules

	return &controller
}

// @Security JWT
// @Summary Create access request
// @Schemes
// @Description Create new access request
// @Tags Access requests
// @Accept json
// @Produce json
// @Param role body models.AccessRequest true "Access request definition"
// @Success 200 {object} ResponseSuccess
// @Router /access/requests [post]
func (r *AccessRequestController) Create(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.RequestController.Create")
	defer span.End()

	uid := c.GetString("uid")
	claims, _ := c.Get("claims")

	data := models.AccessRequest{}
	err := c.ShouldBindBodyWith(&data, binding.JSON)
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorSchemaValidation(err))
		return
	}

	// Retrieve role
	accessRole, err := data.GetRole(r.Roles)
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorSchemaValidation(err))
	}

	// Retrieve approval role
	approvalRule := accessRole.GetApprovalRule(r.ApprovalRules)

	// Modify access request
	data.
		Admit().
		SetStatusPending().
		SetRequester(uid).
		SetTraceId(ctx).
		SetApprovalRule(approvalRule).
		SetExpiration(ctx)

	// Retrieve ProviderUsernames from UserProfile
	profile, err := Db.SelectUserProfile(ctx, models.UserProfile{Id: uid})
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorInvalidUserProfile(err))
		return
	}

	// Validate profile
	err = profile.Validate()
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorInvalidUserProfile(err))
		return
	}
	data.SetProviderUsernames(profile.Settings.ProviderUsernames.ProviderUsernames)

	// ProvideUsernames from Traits should always override UserProfile
	if Config.Auth.JWT.ProviderUsernamesClaim != "" {
		if claimsMap, ok := claims.(models.ClaimsMap); ok {
			usernames := claimsMap.GetProviderUsernamesFromClaim(Config.Auth.JWT.ProviderUsernamesClaim)
			if len(usernames) > 0 {
				data.SetProviderUsernames(usernames)
			}
		}
	}

	// Save it to DB
	Db.InsertAccessRequest(ctx, data)

	// Fire creation event
	Event.AccessRequestCreated(ctx, data)

	c.JSON(errors.StatusCreated())
}

// @Security JWT
// @Summary List access requests
// @Schemes
// @Description List all access requests
// @Tags Access requests
// @Accept json
// @Produce json
// @Success 200 {object} []models.AccessRequest
// @Router /access/requests [get]
func (r *AccessRequestController) List(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.RequestController.List")
	defer span.End()

	uid := c.GetString("uid")
	groups := c.GetStringSlice("groups")
	utype := c.GetString("utype")

	data, err := Db.SelectAccessRequests(ctx)
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseSelect(err))
	}

	// Filter requests to include only those created by the user or those the user has permission to approve
	filtered := []models.AccessRequest{}
	for _, request := range data {
		if request.HasPermissions(uid, groups, utype) || request.Status.RequestedBy == uid {
			filtered = append(filtered, request)
		}
	}

	c.JSON(200, filtered)
}

// @Security JWT
// @Summary Delete access request
// @Schemes
// @Description Delete access request by id
// @Tags Access requests
// @Accept json
// @Produce json
// @Success 200 {object} []models.AccessRequest
// @Router /access/requests/{ID} [delete]
// @Param ID path string true "AccessRequest id" default(xxxx-xxxx-xxxx)
func (r *AccessRequestController) Delete(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.RequestController.Delete")
	defer span.End()

	id := c.Param("ID")
	uid := c.GetString("uid")
	groups := c.GetStringSlice("groups")
	utype := c.GetString("utype")

	accessRequest, err := Db.SelectAccessRequest(ctx, models.AccessRequest{Id: id})
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseRecordNotFound())
		return
	}

	// Check if user is allowed
	allowed := accessRequest.HasPermissions(uid, groups, utype)
	if !allowed {
		c.AbortWithStatusJSON(errors.StatusDenied())
		return
	}

	err = Db.DeleteAccessRequest(ctx, models.AccessRequest{Id: id})
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseSelect(err))
		return
	}

	Event.AccessRequestDeleted(ctx, *accessRequest)

	c.JSON(errors.StatusDeleted())
}

// @Security JWT
// @Summary Approve access request
// @Schemes
// @Description Approve access requests. All providers assigned to role will ensure user access
// @Tags Access requests
// @Accept json
// @Produce json
// @Success 200 {object} ResponseSuccess
// @Router /access/requests/{ID}/approve [post]
// @Param ID path string true "AccessRequest id" default(xxxx-xxxx-xxxx)
func (r *AccessRequestController) Approve(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.RequestController.Approve")
	defer span.End()

	id := c.Param("ID")
	uid := c.GetString("uid")
	groups := c.GetStringSlice("groups")
	utype := c.GetString("utype")

	accessRequest, err := Db.SelectAccessRequest(ctx, models.AccessRequest{Id: id})
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseRecordNotFound())
		return
	}

	// Find role
	accessRole, err := accessRequest.GetRole(r.Roles)
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorSchemaValidation(err))
		return
	}

	// Check if user is allowed
	allowed := accessRequest.HasPermissions(uid, groups, utype)
	if !allowed {
		c.AbortWithStatusJSON(errors.StatusDenied())
		return
	}

	// Call role providers
	err = r.callRoleProvidersAsync(ctx, providerMethodApprove, accessRequest, accessRole)

	// Update request status
	accessRequest.
		SetStatusApprove(uid).
		SetTraceId(ctx)

	Db.UpdateAccessRequest(ctx, accessRequest)

	// Fire approval event
	Event.AccessRequestApproved(ctx, *accessRequest)

	// Partial success
	if err != nil {
		c.AbortWithStatusJSON(errors.AccessProviderCallPartiallyFailed(err))
		return
	}

	c.JSON(errors.StatusUpdated())
}

// @Security JWT
// @Summary Expire access request
// @Schemes
// @Description Expire user access. All providers assigned to role will ensure access expiration
// @Tags Access requests
// @Accept json
// @Produce json
// @Success 200 {object} ResponseSuccess
// @Router /access/requests/{ID}/expire [post]
// @Param ID path string true "AccessRequest id" default(xxxx-xxxx-xxxx)
func (r *AccessRequestController) Expire(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.RequestController.Expire")
	defer span.End()

	id := c.Param("ID")
	uid := c.GetString("uid")
	groups := c.GetStringSlice("groups")
	utype := c.GetString("utype")

	accessRequest, err := Db.SelectAccessRequest(ctx, models.AccessRequest{Id: id})
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseRecordNotFound())
		return
	}

	// Find role
	accessRole, err := accessRequest.GetRole(r.Roles)
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorSchemaValidation(err))
		return
	}

	// Check if user is allowed
	allowed := accessRequest.HasPermissions(uid, groups, utype)
	if !allowed {
		c.AbortWithStatusJSON(errors.StatusDenied())
		return
	}

	// Call role providers
	err = r.callRoleProvidersAsync(ctx, providerMethodExpire, accessRequest, accessRole)

	// Update request status
	accessRequest.
		SetStatusExpired().
		SetTraceId(ctx)

	Db.UpdateAccessRequest(ctx, accessRequest)

	Event.AccessRequestExpired(ctx, *accessRequest)

	if err != nil {
		c.AbortWithStatusJSON(errors.AccessProviderCallPartiallyFailed(err))
		return
	}

	c.JSON(errors.StatusUpdated())
}

type providerMethod int

const (
	providerMethodApprove providerMethod = iota
	providerMethodExpire
)

func (r *AccessRequestController) callRoleProvidersAsync(ctx context.Context, method providerMethod, request *models.AccessRequest, role models.AccessRole) (err error) {

	ctx, span := tracing.NewSpanWrapper(ctx, "controllers.RequestController.callRoleProviders")
	defer span.End()

	runAsync := false

	// Define a WaitGroup to manage concurrent execution
	var wg sync.WaitGroup

	// Use a channel to capture errors from goroutines
	errChan := make(chan error, len(role.Providers))

	// Loop trough providers and initialize them
	for _, config := range role.Providers {

		processProvider := func(config models.ProviderConfig) {

			ctx, span := tracing.NewSpanWrapper(ctx, fmt.Sprintf("controllers.RequestController.callRoleProviders.%s", config.Provider))
			defer span.End()

			config.Parameters["username"] = request.GetProviderUsername(config.Provider)

			provider, err := providers.NewProvider(ctx, config)
			if err != nil {
				request.SetProviderStatusError(config.Name, "NewProvider()", err.Error())
				errChan <- err
				return
			}

			log.Info().
				Str("AccessRequest", request.Id).
				Str("Role", role.Name).
				Str("Provider", config.Name).
				Msg("Calling provider")

			switch method {
			case providerMethodApprove:
				err := provider.GrantAccess(ctx, request)
				if err != nil {
					Event.AccessRequestApprovalError(ctx, *request, config, err)
					errChan <- err
					return
				}
			case providerMethodExpire:
				err := provider.RevokeAccess(ctx, request)
				if err != nil {
					Event.AccessRequestExpireError(ctx, *request, config, err)
					errChan <- err
					return
				}
			default:
				log.Error().Msgf("unknown provider method: %d", method)
			}
		}

		if config.RunAsync {
			// Run asynchronously
			wg.Add(1)
			runAsync = true
			go func(cfg models.ProviderConfig) {
				defer wg.Done()
				processProvider(cfg)
			}(config)
		} else {
			// Run synchronously
			processProvider(config)
		}

	}

	// Wait for all goroutines to complete if running asynchronously
	if runAsync {
		go func() {
			wg.Wait()
			close(errChan)
		}()
	} else {
		close(errChan)
	}

	// Collect errors from the channel
	var errs []error
	for e := range errChan {
		if e != nil {
			errs = append(errs, e)
		}
	}

	// Return combined errors if any
	if len(errs) > 0 {
		return fmt.Errorf("errors occurred: %v", errs)
	}

	return nil
}
