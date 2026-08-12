package middlewares

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Francesco99975/casaintake/internal/helpers"
	"github.com/labstack/echo/v4"
)

func HoneyPotMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if pot := c.FormValue("url"); pot != "" {

				slog.Warn("Honeypot detected", "ip", c.RealIP(), "user-agent", c.Request().UserAgent(), "url", pot)

				return c.NoContent(http.StatusOK)
			}

			return next(c)
		}
	}
}

func MinSubmitTime(min time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.FormValue("start")
			if token == "" {
				// Treat missing token as bot
				slog.Error("Missing token", "ip", c.RealIP(), "user-agent", c.Request().UserAgent())
				return c.NoContent(http.StatusOK) // silent success for bots
			}

			if err := helpers.ValidateStartToken(token, min); err != nil {
				// Log the attempt if you want
				// log.Printf("time trap triggered: %v from %s", err, c.RealIP())
				slog.Warn("Time trap triggered", "ip", c.RealIP(), "user-agent", c.Request().UserAgent(), "error", err)

				// Silent success so bots don't learn
				return c.NoContent(http.StatusOK)
			}

			return next(c)
		}
	}
}
