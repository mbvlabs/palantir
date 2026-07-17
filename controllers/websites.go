package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"palantir/config"
	"palantir/internal/inertia"
	"palantir/internal/storage"
	"palantir/internal/validation"
	"palantir/models"
	"palantir/router"
	"palantir/router/cookies"
	"palantir/router/middleware"
	"palantir/router/routes"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type Websites struct{ db storage.Pool }

func NewWebsites(db storage.Pool) Websites { return Websites{db: db} }

func (w Websites) RegisterRoutes(r *router.Router) error {
	definitions := []echo.Route{
		{Method: http.MethodGet, Path: routes.WebsiteIndex.Path(), Name: routes.WebsiteIndex.Name(), Handler: w.Index},
		{Method: http.MethodGet, Path: routes.WebsiteNew.Path(), Name: routes.WebsiteNew.Name(), Handler: w.New},
		{Method: http.MethodPost, Path: routes.WebsiteCreate.Path(), Name: routes.WebsiteCreate.Name(), Handler: w.Create},
		{Method: http.MethodGet, Path: routes.WebsiteShow.Path(), Name: routes.WebsiteShow.Name(), Handler: w.Show},
		{Method: http.MethodGet, Path: routes.WebsiteEdit.Path(), Name: routes.WebsiteEdit.Name(), Handler: w.Edit},
		{Method: http.MethodPut, Path: routes.WebsiteUpdate.Path(), Name: routes.WebsiteUpdate.Name(), Handler: w.Update},
		{Method: http.MethodDelete, Path: routes.WebsiteDestroy.Path(), Name: routes.WebsiteDestroy.Name(), Handler: w.Destroy},
	}
	var errs []error
	for _, definition := range definitions {
		definition.Middlewares = []echo.MiddlewareFunc{middleware.AuthOnly}
		if _, err := r.AddRoute(definition); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (w Websites) Index(c *echo.Context) error {
	return redirectToSelectedWebsite(c, w.db)
}

func (w Websites) New(c *echo.Context) error {
	websites, err := w.owned(c)
	if err != nil {
		return internalError(c, err)
	}
	return inertia.Page(c, "Websites/New", authenticatedProps(c, inertia.Props{"websites": websites}))
}

func (w Websites) Create(c *echo.Context) error {
	var payload struct{ Name, Domain string }
	if err := c.Bind(&payload); err != nil {
		return badRequest(c)
	}
	ownerID := cookies.ExtractFromCookieApp(c).UserID
	website, err := models.Website.Create(c.Request().Context(), w.db.Executor(), models.CreateWebsiteData{UserID: ownerID, Name: payload.Name, Domain: payload.Domain})
	if err != nil {
		if validationErrors, ok := validation.As(err); ok {
			websites, listErr := w.owned(c)
			if listErr != nil {
				return internalError(c, listErr)
			}
			return inertia.Page(c, "Websites/New", authenticatedProps(c, inertia.Props{"websites": websites}), inertia.WithValidationErrors(validationErrors.ToMap()))
		}
		return internalError(c, err)
	}
	if err := cookies.SetLastWebsite(c, website.ID); err != nil {
		return internalError(c, err)
	}
	return inertia.Redirect(c, routes.WebsiteShow.URL(website.ID), http.StatusSeeOther)
}

func (w Websites) Show(c *echo.Context) error {
	website, err := w.findOwned(c)
	if err != nil {
		return websiteError(c, err)
	}
	websites, err := w.owned(c)
	if err != nil {
		return internalError(c, err)
	}
	return inertia.Page(c, "Websites/Show", authenticatedProps(c, inertia.Props{"website": website, "websites": websites, "trackingScriptURL": config.BaseURL + routes.TrackingScript.URL()}))
}

func (w Websites) Edit(c *echo.Context) error {
	website, err := w.findOwned(c)
	if err != nil {
		return websiteError(c, err)
	}
	websites, err := w.owned(c)
	if err != nil {
		return internalError(c, err)
	}
	return inertia.Page(c, "Websites/Edit", authenticatedProps(c, inertia.Props{"website": website, "websites": websites}))
}

func (w Websites) Update(c *echo.Context) error {
	websiteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return notFound(c)
	}
	var payload struct{ Name, Domain string }
	if err := c.Bind(&payload); err != nil {
		return badRequest(c)
	}
	ownerID := cookies.ExtractFromCookieApp(c).UserID
	website, err := models.Website.UpdateOwned(c.Request().Context(), w.db.Executor(), ownerID, models.UpdateWebsiteData{ID: websiteID, Name: payload.Name, Domain: payload.Domain})
	if err != nil {
		if validationErrors, ok := validation.As(err); ok {
			websites, listErr := w.owned(c)
			if listErr != nil {
				return internalError(c, listErr)
			}
			current, findErr := models.Website.FindOwned(c.Request().Context(), w.db.Executor(), websiteID, ownerID)
			if findErr != nil {
				return websiteError(c, findErr)
			}
			return inertia.Page(c, "Websites/Edit", authenticatedProps(c, inertia.Props{"website": current, "websites": websites}), inertia.WithValidationErrors(validationErrors.ToMap()))
		}
		return websiteError(c, err)
	}
	return inertia.Redirect(c, routes.WebsiteShow.URL(website.ID), http.StatusSeeOther)
}

func (w Websites) Destroy(c *echo.Context) error {
	websiteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return notFound(c)
	}
	ownerID := cookies.ExtractFromCookieApp(c).UserID
	if err := models.Website.DestroyOwned(c.Request().Context(), w.db.Executor(), websiteID, ownerID); err != nil {
		return websiteError(c, err)
	}
	return inertia.Redirect(c, routes.WebsiteIndex.URL(), http.StatusSeeOther)
}

func (w Websites) findOwned(c *echo.Context) (models.WebsiteEntity, error) {
	websiteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return models.WebsiteEntity{}, models.ErrNotFound
	}
	return models.Website.FindOwned(c.Request().Context(), w.db.Executor(), websiteID, cookies.ExtractFromCookieApp(c).UserID)
}

func (w Websites) owned(c *echo.Context) ([]models.WebsiteEntity, error) {
	return models.Website.AllOwned(c.Request().Context(), w.db.Executor(), cookies.ExtractFromCookieApp(c).UserID)
}

func websiteError(c *echo.Context, err error) error {
	if errors.Is(err, models.ErrNotFound) {
		return notFound(c)
	}
	return internalError(c, err)
}

func notFound(c *echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func badRequest(c *echo.Context) error {
	return echo.NewHTTPError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
}

func internalError(c *echo.Context, err error) error {
	slog.ErrorContext(c.Request().Context(), "request failed", "error", err)
	return echo.NewHTTPError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
}
