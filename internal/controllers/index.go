package controllers

import (
	"net/http"

	"github.com/Francesco99975/casaintake/internal/config"
	"github.com/Francesco99975/casaintake/internal/helpers"
	"github.com/Francesco99975/casaintake/views"
	"github.com/labstack/echo/v4"
)

func Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := config.GetDefaultSite(c.Request())

		data.CSRF = c.Get("csrf").(string)
		data.Nonce = c.Get("nonce").(string)

		html := helpers.MustRenderHTML(views.Index(data))

		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}
