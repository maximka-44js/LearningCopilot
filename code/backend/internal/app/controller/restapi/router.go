package restapi

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/maximka-44js/lc/config"
	_ "github.com/maximka-44js/lc/docs" // Swagger docs.
	accessv1 "github.com/maximka-44js/lc/internal/access/controller/restapi/v1"
	"github.com/maximka-44js/lc/internal/app/controller/restapi/middleware"
	projectsv1 "github.com/maximka-44js/lc/internal/projects/controller/restapi/v1"
	storagev1 "github.com/maximka-44js/lc/internal/storage/controller/restapi/v1"
)

// NewRouter -.
// Swagger spec:
//
//	@title       Go Clean Template API
//	@description Multi-domain clean architecture template with translation, user, and task management
//	@version     1.0
//	@host        localhost:8080
//	@BasePath    /v1
//	@securityDefinitions.apikey BearerAuth
//	@in header
//	@name Authorization
func NewRouter(app *fiber.App, cfg *config.Config, deps Deps) {
	// Options
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowCredentials: cfg.CORS.AllowCredentials,
	}))
	app.Use(middleware.Logger(deps.Log))
	app.Use(middleware.Recovery(deps.Log))

	// Prometheus metrics
	if cfg.Metrics.Enabled {
		prometheus := fiberprometheus.New("my-service-name")
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	// Swagger
	if cfg.Swagger.Enabled {
		app.Get("/swagger/*", swagger.New(swagger.Config{
			PersistAuthorization: true,
			OnComplete: template.JS(`function(){
				if(!window.ui || !window.ui.getSystem){return;}
				var authActions = window.ui.getSystem().authActions;
				if(!authActions || !authActions.authorize){return;}
				var original = authActions.authorize;
				authActions.authorize = function(payload){
					if(payload && payload.BearerAuth && payload.BearerAuth.value){
						var value = payload.BearerAuth.value;
						if(value.indexOf("Bearer ") !== 0){
							payload.BearerAuth.value = "Bearer " + value;
						}
					}
					return original(payload);
				};
			}`),
			RequestInterceptor: template.JS(`function(request){
				if(request && request.headers && request.headers.Authorization){
					var auth = request.headers.Authorization;
					if(auth && auth.indexOf("Bearer ") !== 0){
						request.headers.Authorization = "Bearer " + auth;
					}
				}
				return request;
			}`),
		}))
	}

	// K8s probe
	app.Get("/healthz", func(ctx *fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) })
	app.Get("/readyz", func(ctx *fiber.Ctx) error {
		if deps.DB == nil {
			return ctx.SendStatus(http.StatusServiceUnavailable)
		}

		pingCtx, cancel := context.WithTimeout(ctx.Context(), 2*time.Second)
		defer cancel()
		if err := deps.DB.Ping(pingCtx); err != nil {
			return ctx.SendStatus(http.StatusServiceUnavailable)
		}

		return ctx.SendStatus(http.StatusOK)
	})

	// Routers
	apiV1Group := app.Group("/v1")
	{
		storageGroup := apiV1Group.Group("/storage")
		storagev1.NewRoutes(storageGroup, deps.Storage.Library, deps.Storage.Folder, deps.Storage.File, deps.Storage.ObjectStorage, deps.JWT, deps.Access.APIKeys, deps.Log)
		projectsGroup := apiV1Group.Group("/projects")
		projectsv1.NewRoutes(projectsGroup, deps.Projects.Project, deps.Projects.Stage, deps.Projects.Task, deps.Projects.Policy, deps.Projects.ReadModel, deps.JWT, deps.Access.APIKeys, deps.Log)

		accessAPIKeysGroup := apiV1Group.Group("/access/api-keys")
		accessv1.NewRoutes(accessAPIKeysGroup, deps.Access.APIKeys, deps.JWT, deps.Log)

	}
}
