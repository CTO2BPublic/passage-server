package controllers

import (
	"net/http"

	"github.com/CTO2BPublic/passage-server/pkg/models"

	"github.com/gin-gonic/gin"
)

// StatusController handles status probes
type StatusController struct {
}

// NewStatusController creates a new StatusController
func NewStatusController() *StatusController {
	return &StatusController{}
}

// @Security JWT
// @Summary Liveness
// @Schemes
// @Description Liveness
// @Tags API health
// @Accept json
// @Produce json
// @Success 200 {object} []models.Health
// @Router /livez [get]
func (r *StatusController) Livez(c *gin.Context) {

	c.JSON(http.StatusOK, models.Health{
		Healthy: true,
	})
}

// @Security JWT
// @Summary Readyness
// @Schemes
// @Description Readyness
// @Tags API health
// @Accept json
// @Produce json
// @Success 200 {object} []models.Health
// @Router /readyz [get]
func (r *StatusController) Readyz(c *gin.Context) {

	c.JSON(http.StatusOK, models.Health{
		Healthy: true,
	})
}

// @Security JWT
// @Summary Healthy
// @Schemes
// @Description Healthy
// @Tags API health
// @Accept json
// @Produce json
// @Success 200 {object} []models.Health
// @Router /healthz [get]
func (r *StatusController) Healthz(c *gin.Context) {

	c.JSON(http.StatusOK, models.Health{
		Healthy: true,
	})
}
