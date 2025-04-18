package controllers

import (
	"net/http"

	"github.com/CTO2BPublic/passage-server/pkg/errors"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/providers"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// UserController handles user endpoints
type UserController struct {
}

// NewUserController creates a new UserController
func NewUserController() *UserController {
	return &UserController{}
}

// @Security JWT
// @Summary User info
// @Schemes
// @Description Returns information about authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} models.ClaimsMap
// @Router /userinfo [get]
func (r *UserController) UserInfo(c *gin.Context) {

	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"username": "Default user",
		})
		return
	}
	c.JSON(http.StatusOK, claims)
}

// @Security JWT
// @Summary User profile
// @Schemes
// @Description Returns curent user's profile
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} models.UserProfile
// @Router /user/profile [get]
func (r *UserController) GetProfile(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.UserController.GetProfile")
	defer span.End()

	uid := c.GetString("uid")

	exists, err := Db.UserProfileExists(ctx, models.UserProfile{Id: uid})
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseSelect(err))
		return
	}

	// If profile does not exist, create a new one
	if !exists {
		profile := models.UserProfile{
			Id:       uid,
			Username: uid,
			Settings: models.UserProfileSettings{
				ProviderUsernames: providers.NewProviderUsernames(),
			},
		}
		Db.InsertUserProfile(ctx, profile)
		c.JSON(200, profile)
		return
	}

	profile, err := Db.SelectUserProfile(ctx, models.UserProfile{Id: uid})
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseSelect(err))
		return
	}

	c.JSON(200, profile)
}

// @Security JWT
// @Summary User profiles
// @Schemes
// @Description Returns all user profiles
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} []models.User
// @Router /users [get]
func (r *UserController) GetUsers(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.UserController.GetProfiles")
	defer span.End()

	profiles, err := Db.SelectUserProfiles(ctx)
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseSelect(err))
		return
	}

	users := []models.User{}

	for _, profile := range profiles {
		users = append(users, profile.GetUser())
	}

	c.JSON(200, users)
}

// @Security JWT
// @Summary Update user settings
// @Schemes
// @Description Updates current user's settings
// @Tags User
// @Accept json
// @Produce json
// @Param role body models.UserProfileSettings true "User profiles settings"
// @Success 200 {object} ResponseSuccessCreated
// @Router /user/profile/settings [put]
func (r *UserController) UpdateProfileSettings(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.UserController.GetProfile")
	defer span.End()

	uid := c.GetString("uid")

	data := models.UserProfileSettings{}
	err := c.ShouldBindBodyWith(&data, binding.JSON)
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorSchemaValidation(err))
		return
	}

	err = Db.UpdateUserProfile(ctx, models.UserProfile{
		Id:       uid,
		Username: uid,
		Settings: data,
	})
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseInsert(err))
		return
	}

	c.JSON(errors.StatusUpdated())
}
