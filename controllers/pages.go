package controllers

import (
	"errors"
	"net/http"

	"palantir/internal/inertia"
	"palantir/internal/storage"
	"palantir/models"
	"palantir/router"
	"palantir/router/cookies"
	"palantir/router/middleware"
	"palantir/router/routes"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type Pages struct{ db storage.Pool }

func NewPages(db storage.Pool) Pages { return Pages{db: db} }

func (p Pages) RegisterRoutes(r *router.Router) error {
	var errs []error
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		name := routes.HomePage.Name()
		if method == http.MethodHead {
			name += ".head"
		}
		if _, err := r.AddRoute(echo.Route{Method: method, Path: routes.HomePage.Path(), Name: name, Handler: p.Home, Middlewares: []echo.MiddlewareFunc{middleware.AuthOnly}}); err != nil {
			errs = append(errs, err)
		}
	}
	_ = r.AddRouteNotFound(p.NotFound)
	return errors.Join(errs...)
}

func (p Pages) Home(c *echo.Context) error {
	if c.Request().Method == http.MethodHead {
		return c.NoContent(http.StatusOK)
	}
	return redirectToSelectedWebsite(c, p.db)
}

func redirectToSelectedWebsite(c *echo.Context, db storage.Pool) error {
	ownerID := cookies.ExtractFromCookieApp(c).UserID
	lastWebsiteID := cookies.GetLastWebsite(c)
	if lastWebsiteID != uuid.Nil {
		if website, err := models.Website.FindOwned(c.Request().Context(), db.Executor(), lastWebsiteID, ownerID); err == nil {
			return inertia.Redirect(c, routes.WebsiteDashboard.URL(website.ID), http.StatusSeeOther)
		} else if !errors.Is(err, models.ErrNotFound) {
			return internalError(c, err)
		}
	}

	websites, err := models.Website.AllOwned(c.Request().Context(), db.Executor(), ownerID)
	if err != nil {
		return internalError(c, err)
	}
	if len(websites) == 0 {
		return inertia.Redirect(c, routes.WebsiteNew.URL(), http.StatusSeeOther)
	}
	if err := cookies.SetLastWebsite(c, websites[0].ID); err != nil {
		return internalError(c, err)
	}
	return inertia.Redirect(c, routes.WebsiteDashboard.URL(websites[0].ID), http.StatusSeeOther)
}

func (Pages) NotFound(c *echo.Context) error {
	c.Response().WriteHeader(http.StatusNotFound)
	return inertia.Page(c, "Errors/NotFound", inertia.Props{})
}
