package controllers

import (
	"github.com/CTO2BPublic/passage-server/pkg/errors"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/gin-gonic/gin"
)

type ActivityLogController struct {
}

func NewActivityLogController() *ActivityLogController {

	controller := ActivityLogController{}

	return &controller
}

// @Security JWT
// @Summary List events
// @Schemes
// @Description List all events
// @Tags Activity logs
// @Accept json
// @Produce json
// @Success 200 {object} []models.ActivityLog
// @Router /activity-logs [get]
func (r *ActivityLogController) List(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.ActivityLogController.List")
	defer span.End()

	data, err := Db.SelectActivityLogs(ctx)
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseSelect(err))
	}

	c.JSON(200, data)
}
