package controllers

import (
	"github.com/CTO2BPublic/passage-server/pkg/models"

	"github.com/gin-gonic/gin"
)

type AccessRoleController struct {
	Providers map[string]models.ProviderConfig
	Roles     []models.AccessRole
}

func NewAccessRoleController() *AccessRoleController {

	controller := AccessRoleController{}

	// Load roles
	controller.Roles = Config.Roles

	return &controller
}

// @Security JWT
// @Summary Create role
// @Schemes
// @Description Create a new which can be later used in access requests
// @Tags Access roles
// @Accept json
// @Produce json
// @Param role body models.AccessRole true "Role definition"
// @Success 200 {object} ResponseSuccess
// @Router /access/roles [post]
func (r *AccessRoleController) Create(c *gin.Context) {

	c.JSON(200, gin.H{"message": "not implemented. managed via configuration file."})

}

// @Security JWT
// @Summary List roles
// @Schemes
// @Description Create a new which can be later used in access requests
// @Tags Access roles
// @Accept json
// @Produce json
// @Success 200 {object} []models.AccessRole
// @Router /access/roles [get]
func (r *AccessRoleController) List(c *gin.Context) {

	c.JSON(200, r.Roles)

}
