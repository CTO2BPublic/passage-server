package controllers

import (
	"github.com/CTO2BPublic/passage-server/pkg/errors"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/gin-gonic/gin"
)

type EventController struct {
}

func NewEventController() *EventController {

	controller := EventController{}

	return &controller
}

// @Security JWT
// @Summary List events
// @Schemes
// @Description List all events
// @Tags Events
// @Accept json
// @Produce json
// @Success 200 {object} []models.Event
// @Router /events [get]
func (r *EventController) List(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.eventController.List")
	defer span.End()

	data, err := Db.SelectEvents(ctx)
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseSelect(err))
	}

	c.JSON(200, data)
}

// @Security JWT
// @Summary Get event
// @Schemes
// @Description Get single event by id
// @Tags Events
// @Accept json
// @Produce json
// @Success 200 {object} []models.Event
// @Router /events/{ID} [get]
// @Param ID path string true "Event id" default(xxxx-xxxx-xxxx)
func (r *EventController) Get(c *gin.Context) {

	ctx, span := tracing.NewSpanWrapper(c.Request.Context(), "controllers.eventController.Get")
	defer span.End()

	id := c.Param("ID")

	Event, err := Db.SelectEvent(ctx, models.Event{ID: id})
	if err != nil {
		c.AbortWithStatusJSON(errors.ErrorDatabaseRecordNotFound())
		return
	}

	c.JSON(200, Event)
}
