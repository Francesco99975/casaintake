package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Francesco99975/casaintake/cmd/boot"
	"github.com/Francesco99975/casaintake/internal/auth"
	"github.com/Francesco99975/casaintake/internal/config"
	"github.com/Francesco99975/casaintake/internal/enums"
	"github.com/Francesco99975/casaintake/internal/helpers"

	"github.com/Francesco99975/casaintake/internal/controllers"
	"github.com/Francesco99975/casaintake/internal/middlewares"
	"github.com/Francesco99975/casaintake/views"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func createRouter() *echo.Echo {
	e := echo.New()
	e.Logger.SetOutput(io.Discard)
	e.HideBanner = true
	e.HidePort = true
	e.Use(session.Middleware(auth.SessionStore))
	e.Use(middlewares.SlogLogger())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middlewares.RateLimiter())
	// Apply Gzip middleware, but skip it for /metrics
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/metrics" // Skip compression for /metrics
		},
	}))
	e.Use(middlewares.MonitoringMiddleware())
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()), middlewares.MetricsAccessMiddleware())
	e.GET("/healthcheck", func(c echo.Context) error {
		time.Sleep(5 * time.Second)
		return c.JSON(http.StatusOK, "OK")
	})
	e.POST("/csp-violation-report", func(c echo.Context) error {
		type CSPReport struct {
			DocumentURI        string `json:"document-uri"`
			Referrer           string `json:"referrer"`
			ViolatedDirective  string `json:"violated-directive"`
			EffectiveDirective string `json:"effective-directive"`
			OriginalPolicy     string `json:"original-policy"`
			BlockedURI         string `json:"blocked-uri"`
			StatusCode         int    `json:"status-code"`
			SourceFile         string `json:"source-file"`
			LineNumber         int    `json:"line-number"`
			ColumnNumber       int    `json:"column-number"`
		}

		type CSPPayload struct {
			Report CSPReport `json:"csp-report"`
		}

		var payload CSPPayload
		if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
			slog.Warn("CSP Violation Report (unparsable body)", slog.String("err", err.Error()))
			return c.NoContent(http.StatusOK)
		}

		r := payload.Report
		slog.Warn("CSP Violation",
			slog.String("blocked_uri", r.BlockedURI),
			slog.String("violated_directive", r.ViolatedDirective),
			slog.String("effective_directive", r.EffectiveDirective),
			slog.String("document_uri", r.DocumentURI),
			slog.String("source_file", r.SourceFile),
			slog.Int("line", r.LineNumber),
			slog.Int("col", r.ColumnNumber),
			slog.String("original_policy", r.OriginalPolicy),
		)

		return c.NoContent(http.StatusOK)
	})

	e.GET("/sw.js", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/javascript")
		c.Response().Header().Set("Cache-Control", "no-cache")
		return c.File("./static/sw.js")
	})

	e.Static("/assets", "./static")
	e.GET("/assets/dist/*", func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return c.File(filepath.Join("./static/dist", c.Param("*")))
	})

	e.GET("/sitemap.xml", func(c echo.Context) error {
		sitemap := config.GetDefaultSite(c.Request()).Sitemap

		if sitemap == nil {
			slog.Warn("Sitemap not configured")
			return c.NoContent(404)
		}

		return c.Blob(200, "application/xml", sitemap)
	})

	e.GET("/robots.txt", func(c echo.Context) error {
		var content string

		baseURL := boot.Environment.URL

		if boot.Environment.GoEnv == enums.Environments.PRODUCTION {
			content = fmt.Sprintf(`User-agent: *
Disallow: /

Sitemap: %s/sitemap.xml
`, baseURL)
		} else {
			content = fmt.Sprintf(`User-agent: *
Disallow: /

Sitemap: %s/sitemap.xml
`, baseURL)
		}

		return c.Blob(200, "text/plain", []byte(content))
	})

	web := e.Group("")

	web.Use(middlewares.SecurityHeaders())

	web.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "form:_csrf,header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieHTTPOnly: true,
		CookieSecure:   boot.Environment.GoEnv == enums.Environments.PRODUCTION,
		CookieSameSite: http.SameSiteLaxMode,
		Skipper: func(c echo.Context) bool {
			// Skip CSRF for the /webhook route
			return c.Path() == "/webhook"

		},
	}))

	web.GET("/", controllers.Index(), middlewares.GuestMiddleware())
	web.POST("/authorize", controllers.Authorize())
	web.GET("/intake", controllers.Intake(), middlewares.AuthMiddleware())
	web.POST("/intake", controllers.NewPatient(), middlewares.AuthMiddleware())
	web.POST("/logout", controllers.Logout(), middlewares.AuthMiddleware())
	web.GET("/privacy-policy", controllers.PrivacyPolicy())
	web.GET("/terms", controllers.Terms())

	e.HTTPErrorHandler = serverErrorHandler

	return e
}

func serverErrorHandler(err error, c echo.Context) {
	// Default to internal server error (500)
	code := http.StatusInternalServerError
	var message any = "Internal Server Error"

	// Check if it's an echo.HTTPError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message
	}

	// Check the Accept header to decide the response format
	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
		// Respond with JSON if the client prefers JSON
		_ = c.JSON(code, map[string]any{
			"error":   true,
			"message": message,
			"status":  code,
		})
	} else {
		// Prepare data for rendering the error page (HTML)
		data := config.GetDefaultSite(c.Request())
		data.Nonce = c.Get("nonce").(string)
		data.CSRF = c.Get("csrf").(string)

		var title string
		switch code {
		case http.StatusNotFound:
			title = "Not Found"
		case http.StatusUnauthorized:
			title = "Unauthorized"
		case http.StatusForbidden:
			title = "Forbidden"
		case http.StatusInternalServerError:
			title = "Internal Server Error"
		case http.StatusServiceUnavailable:
			title = "Service Unavailable"
		case http.StatusGatewayTimeout:
			title = "Gateway Timeout"
		case http.StatusBadGateway:
			title = "Bad Gateway"
		case http.StatusBadRequest:
			title = "Bad Request"
		case http.StatusConflict:
			title = "Conflict"
		case http.StatusUnprocessableEntity:
			title = "Unprocessable Entity"
		default:
			title = "Internal Server Error"
		}

		var marquee string
		if code >= 500 {
			marquee = "Server Error"
		} else {
			marquee = "Client Error"
		}

		html := helpers.MustRenderHTML(views.Error(data, views.ErrorContentProps{
			Marquee: marquee,
			Code:    fmt.Sprintf("%d", code),
			Title:   title,
			Detail:  message.(string),
		}))

		// Respond with HTML (default) if the client prefers HTML
		_ = c.Blob(code, "text/html; charset=utf-8", html)
	}
}
