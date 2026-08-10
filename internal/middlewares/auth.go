package middlewares

import (
	"net/http"

	"github.com/Francesco99975/casaintake/internal/auth"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authenticated := auth.GetSessionUser(c.Request())
			if !authenticated {
				if c.Request().Header.Get("HX-Request") == "true" {
					c.Response().Header().Set("HX-Redirect", "/")
					return c.NoContent(http.StatusUnauthorized)
				}
				return c.Redirect(http.StatusSeeOther, "/")
			}

			return next(c)
		}
	}
}

func GuestMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if auth.GetSessionUser(c.Request()) {
				if c.Request().Header.Get("HX-Request") == "true" {
					c.Response().Header().Set("HX-Redirect", "/intake")
					return c.NoContent(http.StatusOK)
				}
				return c.Redirect(http.StatusSeeOther, "/intake")
			}
			return next(c)
		}
	}
}
