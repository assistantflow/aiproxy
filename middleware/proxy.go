package middleware

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func Proxy(uri, prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, err := url.Parse(uri)
		if err != nil {
			panic(err)
		}

		p := httputil.NewSingleHostReverseProxy(r)
		p.Director = func(req *http.Request) {
			req.Host = r.Host
			req.URL.Scheme = r.Scheme
			req.URL.Host = r.Host
			tp := filepath.Join("/", prefix)
			if tp != "/" {
				req.URL.Path = strings.TrimPrefix(c.Request.URL.Path, tp)
			}
		}
		p.ServeHTTP(c.Writer, c.Request)

		c.Abort()
	}
}
