package router

import (
	"errors"
	"net/http"

	"palantir/controllers"
	"palantir/router/routes"

	"github.com/labstack/echo/v5"
)

func (r Router) RegisterPagesRoutes(pages controllers.Pages) error {
	errs := []error{}

	_, err := r.e.AddRoute(echo.Route{
		Method:  http.MethodGet,
		Path:    routes.HomePage.Path(),
		Name:    routes.HomePage.Name(),
		Handler: pages.Home,
	})
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
