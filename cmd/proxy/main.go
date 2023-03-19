package main

import (
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/douglarek/ai-proxy/middleware"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	aiURI      = flag.String("ai_url", "https://api.openai.com", "the uri to proxy")
	prefixPath = flag.String("prefix", "/", "the proxy prefix")
	port       = flag.Int("port", 8080, "the proxy server port")
	metricPort = flag.Int("metric_port", 2112, "the metric server port")
	debug      = flag.Bool("debug", false, "debug mode")
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func startProm() {
	log.Info().Msgf("start metric server on port %d", *metricPort)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *metricPort), nil); err != nil {
		log.Panic().Err(err).Msg("failed to start metric server")
	}
}

func main() {
	flag.Parse()
	if *aiURI == "" {
		log.Panic().Msg("no ai_url provided")
	}
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(gin.Recovery())
	r.Use(logger.SetLogger(logger.WithLogger(func(*gin.Context, zerolog.Logger) zerolog.Logger {
		return log.Logger.With().Str("app", "ai-proxy").Logger()
	})))

	g := r.Group(filepath.Join("/", *prefixPath))
	rp := middleware.Proxy(*aiURI, *prefixPath)
	g.Any("*p", rp)

	go startProm()

	log.Info().Msgf("start server on port %d", *port)
	if err := r.Run(fmt.Sprintf(":%d", *port)); err != nil {
		log.Error().Err(err).Msg("failed to start proxy")
	}
}
