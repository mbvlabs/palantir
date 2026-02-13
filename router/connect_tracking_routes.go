package router

import (
	"errors"
	"net/http"

	"palantir/controllers"
	"palantir/router/routes"

	"github.com/labstack/echo/v5"
)

func (r Router) RegisterTrackingRoutes(tracking controllers.Tracking) error {
	errs := []error{}

	_, err := r.e.AddRoute(echo.Route{
		Method:  http.MethodGet,
		Path:    routes.TrackingScript.Path(),
		Name:    routes.TrackingScript.Name(),
		Handler: tracking.Script,
	})
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
