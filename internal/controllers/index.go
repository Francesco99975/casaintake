package controllers

import (
	"net/http"

	"github.com/Francesco99975/casaintake/cmd/boot"
	"github.com/Francesco99975/casaintake/internal/config"
	"github.com/Francesco99975/casaintake/internal/enums"
	"github.com/Francesco99975/casaintake/internal/helpers"
	"github.com/Francesco99975/casaintake/internal/httperr"
	"github.com/Francesco99975/casaintake/internal/models"
	"github.com/Francesco99975/casaintake/internal/tools"
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

func Intake() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := config.GetDefaultSite(c.Request())

		if csrf, ok := c.Get("csrf").(string); ok {
			data.CSRF = csrf
		} else {
			// log it so you can see it in the server console
			c.Logger().Error("csrf token missing from context")
			return echo.NewHTTPError(http.StatusInternalServerError, "csrf missing")
		}
		data.Nonce = c.Get("nonce").(string)

		html := helpers.MustRenderHTML(views.PatientIntakeForm(data))

		return c.Blob(http.StatusOK, "text/html; charset=utf-8", html)
	}
}

func NewPatient() echo.HandlerFunc {
	return func(c echo.Context) error {
		herr := httperr.New("intake", "NewPatient", c.Request().Header.Get("X-Request-ID"))

		var payload models.PatientIntakeRequest
		if err := c.Bind(&payload); err != nil {
			return herr.Handle(c.Response(), http.StatusBadRequest, err)
		}

		err := payload.Validate()
		if err != nil {
			return herr.Handle(c.Response(), http.StatusBadRequest, err)
		}

		go helpers.ResendNewPatientTemplate(boot.Environment.RecipientEmail, payload)

		tools.SetToastTrigger(c.Response(), enums.SuccessToast, "Patient intake submitted successfully")

		return c.NoContent(http.StatusOK)
	}
}
