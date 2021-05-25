package fiber

import (
	"log"
	"net/http"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	DefaultEndpointHealth  = "healthz"
	DefaultEndpointMetrics = "metrics"
	DefaultIdleTimeout     = 10 * time.Second
	DefaultReadTimeout     = 10 * time.Second
	DefaultWriteTimeout    = 10 * time.Second
)

var (
	EndpointHealth  = DefaultEndpointHealth
	EndpointMetrics = DefaultEndpointMetrics
	DefaultConfig   = fiber.Config{
		IdleTimeout:  DefaultIdleTimeout,
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
	}
)

func New(addr string, configs ...fiber.Config) (server *fiber.App, listen func() error, shutdown func()) {
	if len(configs) == 0 {
		configs = []fiber.Config{DefaultConfig}
	}

	server = fiber.New(configs...)
	UseAll(server)

	listen = func() error {
		return server.Listen(addr)
	}

	shutdown = func() {
		if err := server.Shutdown(); err != nil {
			log.Printf("Fail to shutdown server: %s", err)
		}
	}
	return
}

func UseAll(server fiber.Router) {
	server.Use(EndpointMetrics, Metrics())
	server.Use(EndpointHealth, Health())
	server.Use(logger.New())
}

func Metrics() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.Handler())
}

func Health() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	}
}
