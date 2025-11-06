package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

// NOTE: /api/svedprint/* -> svedprint
//       /api/admin/*     -> svedprint-admin
//       /api/print/*     -> svedprint-print

func configureProxies(gs *GinServer) {
	targets := map[string]string{
		"svedprint": os.Getenv("SVEDPRINT_SERVICE_URL"),
		"admin":     os.Getenv("SVEDPRINT_ADMIN_SERVICE_URL"),
		"print":     os.Getenv("SVEDPRINT_PRINT_SERVICE_URL"),
	}

	for groupName, targetUrl := range targets {
		target, _ := url.Parse(targetUrl)
		proxy := httputil.NewSingleHostReverseProxy(target)

		originalDirector := proxy.Director
		proxy.Director = func(r *http.Request) {
			originalDirector(r)
			r.URL.Host = target.Host
			r.URL.Scheme = target.Scheme
			r.Host = target.Host
			r.Header.Set("X-Forwarded-Host", r.Host) // TODO: Maybe add roles and username?
		}

		gs.proxies[groupName] = proxy
	}
}

func (gs *GinServer) ProxyTo(groupName string) gin.HandlerFunc {
	proxy := gs.proxies[groupName]

	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
