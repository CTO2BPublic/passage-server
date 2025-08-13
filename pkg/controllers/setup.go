package controllers

import (
	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/dbdriver"
	"github.com/CTO2BPublic/passage-server/pkg/eventdriver"
	"go.opentelemetry.io/otel"
)

var Config = config.GetConfig()
var Db = dbdriver.GetDriver()
var Event = eventdriver.GetDriver()
var Tracer = otel.Tracer("pkg/controllers/requestController")
