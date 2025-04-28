package router

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/raikh/calc_micro_final/controller"
	"github.com/raikh/calc_micro_final/internal/config"
	"github.com/raikh/calc_micro_final/middleware"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func InitRouter(cfg *config.Config) error {
	httpAddr := fmt.Sprintf("%s:%s", cfg.GetKey("APP_LISTENING_ADDRESS"), cfg.GetKey("APP_HTTP_LISTEN_PORT"))
	log.Printf("HTTP server listening on %s", httpAddr)
	middleware.Init(cfg)

	e := echo.New()
	e.Debug = true
	if cfg.GetKey("APP_ENV") == "production" {
		e.HideBanner = true
		e.HidePort = true
		e.Debug = false
	}

	e.Use(echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values echoMiddleware.RequestLoggerValues) error {
			log.WithFields(logrus.Fields{
				"URI":    values.URI,
				"Status": values.Status,
			}).Info("request")

			return nil
		},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/api/register", controller.Register())
	e.POST("/api/login", controller.Login())

	apiGroup := e.Group("/api")
	apiGroup.Use(middleware.JwtAuthMiddleware)
	apiGroup.Add(http.MethodPost, "/calculate", controller.HandleCalculate(buildDelayDict(cfg)))
	apiGroup.Add(http.MethodGet, "/expressions", controller.HandleGetExpressions())
	apiGroup.Add(http.MethodGet, "/expressions/:id", controller.HandleGetExpressionsById())

	return e.Start(httpAddr)
}

func buildDelayDict(cfg *config.Config) map[string]int64 {
	delayDict := make(map[string]int64)
	delayDict["+"] = strToInt64(cfg.GetKey("TIME_ADDITION_MS"))
	delayDict["-"] = strToInt64(cfg.GetKey("TIME_SUBTRACTION_MS"))
	delayDict["*"] = strToInt64(cfg.GetKey("TIME_MULTIPLICATIONS_MS"))
	delayDict["/"] = strToInt64(cfg.GetKey("TIME_DIVISIONS_MS"))
	return delayDict
}

func strToInt64(str string) int64 {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		fmt.Println("Error converting string to int64:", err)
		return 0
	}
	return value
}
