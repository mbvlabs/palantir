package router

import (
	"errors"
	"net/http"

	"palantir/controllers"
	"palantir/router/middleware"
	"palantir/router/routes"

	"github.com/labstack/echo/v5"
)

func (r Router) RegisterDashboardRoutes(dashboard controllers.Dashboard) error {
	errs := []error{}

	_, err := r.e.AddRoute(echo.Route{
		Method:  http.MethodGet,
		Path:    routes.WebsiteDashboard.Path(),
		Name:    routes.WebsiteDashboard.Name(),
		Handler: dashboard.Show,
		Middlewares: []echo.MiddlewareFunc{
			middleware.AuthOnly,
		},
	})
	if err != nil {
		errs = append(errs, err)
	}

	_, err = r.e.AddRoute(echo.Route{
		Method:  http.MethodGet,
		Path:    routes.WebsiteDashboardLive.Path(),
		Name:    routes.WebsiteDashboardLive.Name(),
		Handler: dashboard.Live,
		Middlewares: []echo.MiddlewareFunc{
			middleware.AuthOnly,
		},
	})
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
