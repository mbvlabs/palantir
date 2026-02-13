package router

import (
	"errors"
	"net/http"

	"palantir/controllers"
	"palantir/router/routes"

	"github.com/labstack/echo/v5"
)

func (r Router) RegisterCollectRoutes(collect controllers.Collect) error {
	errs := []error{}

	_, err := r.e.AddRoute(echo.Route{
		Method:  http.MethodOptions,
		Path:    routes.CollectCreate.Path(),
		Name:    routes.CollectCreate.Name() + ".options",
		Handler: collect.Preflight,
	})
	if err != nil {
		errs = append(errs, err)
	}

	_, err = r.e.AddRoute(echo.Route{
		Method:  http.MethodPost,
		Path:    routes.CollectCreate.Path(),
		Name:    routes.CollectCreate.Name(),
		Handler: collect.Create,
	})
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
