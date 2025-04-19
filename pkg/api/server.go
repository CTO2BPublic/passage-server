// pkg/api/server.go
package api

import (
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/CTO2BPublic/passage-server/pkg/controllers"
	"github.com/CTO2BPublic/passage-server/pkg/dbdriver"
	"github.com/CTO2BPublic/passage-server/pkg/eventdriver"
	"github.com/CTO2BPublic/passage-server/pkg/middlewares"

	docs "github.com/CTO2BPublic/passage-server/docs"

	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var Db = dbdriver.GetDriver()
var Config = config.GetConfig()
var Event = eventdriver.GetDriver()

type Server struct {
	Engine *gin.Engine
}

func (s *Server) SetupEngineWithDefaults() *Server {

	s.Engine = gin.New()

	if Config.Tracing.Enabled {
		tracing.NewTracer()
		s.Engine.Use(tracing.NewTracingMidleware())
	}

	if Config.Events.Kafka.Enabled {
		Event.NewDriver()
	}

	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	s.Engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz", "/readyz", "/livez"},
	}))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	s.Engine.Use(gin.Recovery())

	// Middleware CORS
	s.Engine.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"*"},
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin", "Content-type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Total-Count", "Link"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API documentation
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Host = Config.Swagger.Host
	docs.SwaggerInfo.Schemes = []string{"https", "http"}

	// Initialize drivers
	if Config.Db.Engine != "" {
		Db.Connect()
		Db.AutoMigrate()
	}

	// Initialize controllers
	accessRoleController := controllers.NewAccessRoleController()
	accessRequestController := controllers.NewAccessRequestController()
	statusController := controllers.NewStatusController()
	userController := controllers.NewUserController()

	// Define routes
	rg := s.Engine.Group("")
	access := rg.Group("/access")
	access.Use(middlewares.Auth())
	{
		access.POST("/roles", accessRoleController.Create)
		access.GET("/roles", accessRoleController.List)
		access.POST("/requests", accessRequestController.Create)
		access.GET("/requests", accessRequestController.List)
		access.POST("/requests/:ID/approve", accessRequestController.Approve)
		access.POST("/requests/:ID/expire", accessRequestController.Expire)
		access.DELETE("/requests/:ID", accessRequestController.Delete)
	}

	user := rg.Group("/user")
	user.Use(middlewares.Auth())
	{
		user.GET("/profile", userController.GetProfile)
		user.PUT("/profile/settings", userController.UpdateProfileSettings)
	}

	users := rg.Group("/users")
	users.Use(middlewares.Auth())
	{
		users.GET("/", userController.GetUsers)
	}

	s.Engine.GET("/userinfo", middlewares.Auth(), userController.UserInfo)

	s.Engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	s.Engine.GET("/healthz", statusController.Healthz)
	s.Engine.GET("/readyz", statusController.Readyz)
	s.Engine.GET("/livez", statusController.Livez)

	return s
}

func (s *Server) RunEngine() {
	err := s.Engine.Run()
	if err != nil {
		log.Error().Str("Engine", "Failed to start apiserver").Msg(err.Error())
	}
}

func GetServer() *Server {
	return new(Server)
}
