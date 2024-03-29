package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

var tokenUsed = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "token_used_total",
		Help: "How many token used, partitioned by api key.",
	},
	[]string{"key"},
)

func init() {
	prometheus.Register(tokenUsed)
}

func modifyResponse(key string, stream bool) func(*http.Response) error {
	return func(r *http.Response) error {
		if r.StatusCode == http.StatusOK {
			if !stream {
				var b bytes.Buffer
				cb := io.TeeReader(r.Body, &b)
				var s struct {
					Usage struct {
						PromptTokens     int `json:"prompt_tokens"`
						CompletionTokens int `json:"completion_tokens"`
						TotalTokens      int `json:"total_tokens"`
					} `json:"usage"`
				}
				err := json.NewDecoder(cb).Decode(&s)
				if err != nil {
					log.Error().Err(err).Msg("decode proxy response failed")
				} else {
					tokenUsed.WithLabelValues(key).Add(float64(s.Usage.TotalTokens))
					log.Debug().Str("key", key).Interface("usage", s).Send()
				}
				r.Body = io.NopCloser(&b)
			}
		}
		return nil
	}
}

func parseAuth(s string) string {
	bk := strings.Split(strings.TrimSpace(s), " ")
	var k string
	if len(bk) > 0 {
		k = bk[len(bk)-1]
	}
	return k
}

func Proxy(uri, prefix string) gin.HandlerFunc {
	r, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	p := httputil.NewSingleHostReverseProxy(r)
	return func(c *gin.Context) {
		var b bytes.Buffer
		cb := io.TeeReader(c.Request.Body, &b)
		var d struct {
			Stream bool `json:"stream"`
		}
		if err := json.NewDecoder(cb).Decode(&d); err != nil && err != io.EOF { // if no request body, err will be EOF
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		c.Request.Body = io.NopCloser(&b)
		p.Director = func(req *http.Request) {
			req.Host = r.Host
			req.URL.Scheme = r.Scheme
			req.URL.Host = r.Host
			tp := filepath.Join("/", prefix)
			if tp != "/" {
				req.URL.Path = strings.TrimPrefix(c.Request.URL.Path, tp)
			}
		}
		p.ModifyResponse = modifyResponse(parseAuth(c.GetHeader("Authorization")), d.Stream)
		p.ServeHTTP(c.Writer, c.Request)
		c.Next()
	}
}
