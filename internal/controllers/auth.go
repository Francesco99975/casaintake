package controllers

import (
	"errors"
	"net/http"

	"github.com/Francesco99975/casaintake/cmd/boot"
	"github.com/Francesco99975/casaintake/internal/auth"
	"github.com/Francesco99975/casaintake/internal/helpers"
	"github.com/Francesco99975/casaintake/internal/httperr"
	"github.com/labstack/echo/v4"
)

func Authorize() echo.HandlerFunc {
	return func(c echo.Context) error {
		herr := httperr.New("authenticating for authorization", "Authorize", c.Request().Header.Get("X-Request-ID"))
		passkey := c.FormValue("passkey")

		if !helpers.CheckPasswordHash(passkey, boot.Environment.PassKey) {
			return herr.Why("invalid password").Handle(c.Response(), http.StatusUnauthorized, errors.New("invalid password"))
		}

		err := auth.SetSessionUser(c.Response(), c.Request())
		if err != nil {
			return herr.Why("failed to set session user").Handle(c.Response(), http.StatusInternalServerError, err)
		}

		c.Response().Header().Set("HX-Redirect", "/intake")
		return c.NoContent(http.StatusSeeOther)
	}
}

func Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		err := auth.ClearSession(c.Response(), c.Request())
		if err != nil {
			return err
		}

		c.Response().Header().Set("HX-Redirect", "/")
		return c.NoContent(http.StatusSeeOther)
	}
}
