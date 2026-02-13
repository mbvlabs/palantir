package router

import (
	"errors"
	"net/http"

	"palantir/controllers"
	"palantir/router/middleware"
	"palantir/router/routes"

	"github.com/labstack/echo/v5"
)

func (r Router) RegisterWebsitesRoutes(websites controllers.Websites) error {
	errs := []error{}

	_, err := r.e.AddRoute(echo.Route{
		Method:  http.MethodGet,
		Path:    routes.WebsiteIndex.Path(),
		Name:    routes.WebsiteIndex.Name(),
		Handler: websites.Index,
		Middlewares: []echo.MiddlewareFunc{
			middleware.AuthOnly,
		},
	})
	if err != nil {
		errs = append(errs, err)
	}

	_, err = r.e.AddRoute(echo.Route{
		Method:  http.MethodGet,
		Path:    routes.WebsiteNew.Path(),
		Name:    routes.WebsiteNew.Name(),
		Handler: websites.New,
		Middlewares: []echo.MiddlewareFunc{
			middleware.AuthOnly,
		},
	})
	if err != nil {
		errs = append(errs, err)
	}

	_, err = r.e.AddRoute(echo.Route{
		Method:  http.MethodPost,
		Path:    routes.WebsiteCreate.Path(),
		Name:    routes.WebsiteCreate.Name(),
		Handler: websites.Create,
		Middlewares: []echo.MiddlewareFunc{
			middleware.AuthOnly,
		},
	})
	if err != nil {
		errs = append(errs, err)
	}

	_, err = r.e.AddRoute(echo.Route{
		Method:  http.MethodGet,
		Path:    routes.WebsiteShow.Path(),
		Name:    routes.WebsiteShow.Name(),
		Handler: websites.Show,
		Middlewares: []echo.MiddlewareFunc{
			middleware.AuthOnly,
		},
	})
	if err != nil {
		errs = append(errs, err)
	}

	_, err = r.e.AddRoute(echo.Route{
		Method:  http.MethodGet,
		Path:    routes.WebsiteEdit.Path(),
		Name:    routes.WebsiteEdit.Name(),
		Handler: websites.Edit,
		Middlewares: []echo.MiddlewareFunc{
			middleware.AuthOnly,
		},
	})
	if err != nil {
		errs = append(errs, err)
	}

	_, err = r.e.AddRoute(echo.Route{
		Method:  http.MethodPut,
		Path:    routes.WebsiteUpdate.Path(),
		Name:    routes.WebsiteUpdate.Name(),
		Handler: websites.Update,
		Middlewares: []echo.MiddlewareFunc{
			middleware.AuthOnly,
		},
	})
	if err != nil {
		errs = append(errs, err)
	}

	_, err = r.e.AddRoute(echo.Route{
		Method:  http.MethodDelete,
		Path:    routes.WebsiteDestroy.Path(),
		Name:    routes.WebsiteDestroy.Name(),
		Handler: websites.Destroy,
		Middlewares: []echo.MiddlewareFunc{
			middleware.AuthOnly,
		},
	})
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
