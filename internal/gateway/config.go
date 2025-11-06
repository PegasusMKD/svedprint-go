package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/PegasusMKD/svedprint-go/internal/gateway/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
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
		start := time.Now()
		incomingPath := c.Request.URL.Path
		wrappedWriter := &responseWriter{
			ResponseWriter: c.Writer,
			statusCode:     200,
		}
		c.Writer = wrappedWriter
		strippedPath := c.Param("path")
		c.Request.URL.Path = strippedPath

		proxy.ServeHTTP(c.Writer, c.Request)

		var errorMsg pgtype.Text
		if wrappedWriter.statusCode >= 400 {
			errorMsg = pgtype.Text{
				// TODO: Rewrite so we parse the actual error message
				String: fmt.Sprintf("Request failed with status %d", wrappedWriter.statusCode),
				Valid:  true,
			}
		}

		logEntry := sqlc.RequestLog{
			Timestamp: pgtype.Timestamptz{
				Time:  time.Now(),
				Valid: true,
			},
			Method:       c.Request.Method,
			IncomingPath: incomingPath,
			RedirectedPath: pgtype.Text{
				String: c.Request.URL.Path,
				Valid:  true,
			},
			// TODO: Pull from Auth middleware
			OrganizationID: pgtype.Text{
				String: "organization",
				Valid:  true,
			},
			UserID: pgtype.Text{
				String: "user",
				Valid:  true,
			},
			StatusCode:     int32(wrappedWriter.statusCode),
			ResponseTimeMs: int32(time.Since(start).Milliseconds()),
			UpstreamService: pgtype.Text{
				String: groupName,
				Valid:  true,
			},
			ErrorMessage: errorMsg,
		}

		gs.requestLogWriter.Write(logEntry)
	}
}
