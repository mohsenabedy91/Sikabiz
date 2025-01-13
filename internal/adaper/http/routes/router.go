package routes

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/http/middlewares"
	"github.com/mohsenabedy91/Sikabiz/internal/core/config"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
	"github.com/mohsenabedy91/Sikabiz/pkg/metrics"
	"github.com/mohsenabedy91/Sikabiz/pkg/translation"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Router is a wrapper for HTTP router
type Router struct {
	Engine *gin.Engine
	log    logger.Logger
	conf   config.Config
	trans  translation.Translator
}

// NewRouter creates a new HTTP router
func NewRouter(
	log logger.Logger,
	conf config.Config,
	trans translation.Translator,
) (*Router, error) {

	// Disable debug mode in production
	if conf.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	RegisterPrometheus(log)

	router.Use(middlewares.Prometheus())
	router.Use(gin.Logger(), gin.CustomRecovery(middlewares.ErrorHandler(trans)))
	router.Use(middlewares.DefaultStructuredLogger(log))

	setSwaggerRoutes(router.Group(""), conf.Swagger)

	router.GET("metrics", gin.WrapH(promhttp.Handler()))

	return &Router{
		Engine: router,
		log:    log,
		conf:   conf,
		trans:  trans,
	}, nil
}

// Serve starts the HTTP server
func (r *Router) Serve(server *http.Server) {
	go func() {
		// service connections
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			r.log.Error(logger.Internal, logger.Startup, fmt.Sprintf("Error starting the HTTP server: %v", err), nil)
		}
	}()
}

func RegisterPrometheus(log logger.Logger) {
	err := prometheus.Register(metrics.DbCall)
	if err != nil {
		log.Error(logger.Prometheus, logger.Startup, err.Error(), nil)
	}

	err = prometheus.Register(metrics.HttpDuration)
	if err != nil {
		log.Error(logger.Prometheus, logger.Startup, err.Error(), nil)
	}
}
