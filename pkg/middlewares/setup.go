package middlewares

import (
	"github.com/CTO2BPublic/passage-server/pkg/config"

	"go.opentelemetry.io/otel"
)

var Config = config.GetConfig()
var Tracer = otel.Tracer("pkg/middlewares")
