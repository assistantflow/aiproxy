package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	aiURI      = flag.String("ai_url", "https://api.openai.com", "the uri to proxy")
	prefixPath = flag.String("prefix", "/", "the proxy prefix")
	port       = flag.Int("port", 8080, "the server port")
)

func reverseOpenAI(c *gin.Context) {
	r, err := url.Parse(*aiURI)
	if err != nil {
		panic(err)
	}

	p := httputil.NewSingleHostReverseProxy(r)
	p.Director = func(req *http.Request) {
		req.Host = r.Host
		req.URL.Scheme = r.Scheme
		req.URL.Host = r.Host
		tp := filepath.Join("/", *prefixPath)
		if tp != "/" {
			req.URL.Path = strings.TrimPrefix(c.Request.URL.Path, tp)
		}
	}
	p.ServeHTTP(c.Writer, c.Request)

	c.Abort()
}

func main() {
	flag.Parse()
	if *aiURI == "" {
		log.Panic().Msg("no ai_url provided")
	}

	r := gin.Default()
	{
		g := r.Group(filepath.Join("/", *prefixPath))
		g.GET("*p", reverseOpenAI)
		g.POST("*p", reverseOpenAI)
		g.DELETE("*p", reverseOpenAI)
	}
	if err := r.Run(fmt.Sprintf(":%d", *port)); err != nil {
		log.Error().Err(err).Msg("failed to start proxy")
	}
}
