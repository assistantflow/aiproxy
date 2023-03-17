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

func ProxyFunc(uri, prefix string) http.HandlerFunc {
	remoteURL, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	p := httputil.NewSingleHostReverseProxy(remoteURL)
	return func(w http.ResponseWriter, r *http.Request) {
		var b bytes.Buffer
		cb := io.TeeReader(r.Body, &b)
		var d struct {
			Stream bool `json:"stream"`
		}
		if err := json.NewDecoder(cb).Decode(&d); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(&b)
		p.Director = func(req *http.Request) {
			req.Host = remoteURL.Host
			req.URL.Scheme = remoteURL.Scheme
			req.URL.Host = remoteURL.Host
			tp := filepath.Join("/", prefix)
			if tp != "/" {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, tp)
			}
		}
		p.ModifyResponse = modifyResponse(parseAuth(r.Header.Get("Authorization")), d.Stream)
		p.ServeHTTP(w, r)
	}
}

func Proxy(uri, prefix string) gin.HandlerFunc {
	return gin.WrapF(ProxyFunc(uri, prefix))
}
