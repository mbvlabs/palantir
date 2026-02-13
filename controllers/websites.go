package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"palantir/internal/storage"
	"palantir/models"
	"palantir/router/cookies"
	"palantir/router/routes"
	"palantir/views"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type Websites struct {
	db storage.Pool
}

func NewWebsites(db storage.Pool) Websites {
	return Websites{db: db}
}

func (w Websites) Index(etx *echo.Context) error {
	app := cookies.GetApp(etx)
	ctx := etx.Request().Context()

	websites, err := models.FindWebsitesByUserID(ctx, w.db.Conn(), app.UserID)
	if err != nil {
		return render(etx, views.InternalError())
	}

	return render(etx, views.WebsitesIndex(websites))
}

func (w Websites) New(etx *echo.Context) error {
	return render(etx, views.WebsitesNew())
}

type createWebsitePayload struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

func (w Websites) Create(etx *echo.Context) error {
	var payload createWebsitePayload
	if err := etx.Bind(&payload); err != nil {
		return render(etx, views.BadRequest())
	}

	app := cookies.GetApp(etx)
	ctx := etx.Request().Context()

	website, err := models.CreateWebsite(ctx, w.db.Conn(), models.CreateWebsiteData{
		UserID: app.UserID,
		Name:   payload.Name,
		Domain: payload.Domain,
	})
	if err != nil {
		if errors.Is(err, models.ErrDomainValidation) {
			cookies.AddFlash(etx, cookies.FlashError, "Please provide a valid name and domain")
			return etx.Redirect(http.StatusSeeOther, routes.WebsiteNew.URL())
		}
		return render(etx, views.InternalError())
	}

	cookies.AddFlash(etx, cookies.FlashSuccess, "Website added successfully")
	return etx.Redirect(http.StatusSeeOther, routes.WebsiteShow.URL(website.ID))
}

func (w Websites) Show(etx *echo.Context) error {
	websiteID, err := uuid.Parse(etx.Param("id"))
	if err != nil {
		return render(etx, views.BadRequest())
	}

	app := cookies.GetApp(etx)
	ctx := etx.Request().Context()

	website, err := models.FindWebsite(ctx, w.db.Conn(), websiteID)
	if err != nil {
		return render(etx, views.NotFound())
	}

	if website.UserID != app.UserID {
		return render(etx, views.NotFound())
	}

	return render(etx, views.WebsitesShow(website))
}

func (w Websites) Edit(etx *echo.Context) error {
	websiteID, err := uuid.Parse(etx.Param("id"))
	if err != nil {
		return render(etx, views.BadRequest())
	}

	app := cookies.GetApp(etx)
	ctx := etx.Request().Context()

	website, err := models.FindWebsite(ctx, w.db.Conn(), websiteID)
	if err != nil {
		return render(etx, views.NotFound())
	}

	if website.UserID != app.UserID {
		return render(etx, views.NotFound())
	}

	return render(etx, views.WebsitesEdit(website))
}

type updateWebsitePayload struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

func (w Websites) Update(etx *echo.Context) error {
	websiteID, err := uuid.Parse(etx.Param("id"))
	if err != nil {
		return render(etx, views.BadRequest())
	}

	app := cookies.GetApp(etx)
	ctx := etx.Request().Context()

	website, err := models.FindWebsite(ctx, w.db.Conn(), websiteID)
	if err != nil {
		return render(etx, views.NotFound())
	}

	if website.UserID != app.UserID {
		return render(etx, views.NotFound())
	}

	var payload updateWebsitePayload
	if err := etx.Bind(&payload); err != nil {
		return render(etx, views.BadRequest())
	}

	_, err = models.UpdateWebsite(ctx, w.db.Conn(), models.UpdateWebsiteData{
		ID:     websiteID,
		Name:   payload.Name,
		Domain: payload.Domain,
	})
	if err != nil {
		if errors.Is(err, models.ErrDomainValidation) {
			cookies.AddFlash(etx, cookies.FlashError, "Please provide a valid name and domain")
			return etx.Redirect(http.StatusSeeOther, routes.WebsiteEdit.URL(websiteID))
		}
		return render(etx, views.InternalError())
	}

	cookies.AddFlash(etx, cookies.FlashSuccess, "Website updated successfully")
	return etx.Redirect(http.StatusSeeOther, routes.WebsiteShow.URL(websiteID))
}

func (w Websites) Destroy(etx *echo.Context) error {
	websiteID, err := uuid.Parse(etx.Param("id"))
	if err != nil {
		return render(etx, views.BadRequest())
	}

	app := cookies.GetApp(etx)
	ctx := etx.Request().Context()

	website, err := models.FindWebsite(ctx, w.db.Conn(), websiteID)
	if err != nil {
		return render(etx, views.NotFound())
	}

	if website.UserID != app.UserID {
		return render(etx, views.NotFound())
	}

	if err := models.DestroyWebsite(ctx, w.db.Conn(), websiteID); err != nil {
		cookies.AddFlash(etx, cookies.FlashError, fmt.Sprintf("Failed to delete website: %v", err))
		return etx.Redirect(http.StatusSeeOther, routes.WebsiteShow.URL(websiteID))
	}

	cookies.AddFlash(etx, cookies.FlashSuccess, "Website deleted successfully")
	return etx.Redirect(http.StatusSeeOther, routes.WebsiteIndex.URL())
}
