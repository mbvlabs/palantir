package routes

import "palantir/internal/routing"

const WebsitesPrefix = "/websites"

var (
	WebsiteIndex     = routing.NewSimpleRoute("", "websites.index", WebsitesPrefix)
	WebsiteNew       = routing.NewSimpleRoute("/new", "websites.new", WebsitesPrefix)
	WebsiteCreate    = routing.NewSimpleRoute("", "websites.create", WebsitesPrefix)
	WebsiteShow      = routing.NewRouteWithUUIDID("/:id", "websites.show", WebsitesPrefix)
	WebsiteEdit      = routing.NewRouteWithUUIDID("/:id/edit", "websites.edit", WebsitesPrefix)
	WebsiteUpdate    = routing.NewRouteWithUUIDID("/:id", "websites.update", WebsitesPrefix)
	WebsiteDestroy   = routing.NewRouteWithUUIDID("/:id", "websites.destroy", WebsitesPrefix)
	WebsiteDashboard = routing.NewRouteWithUUIDID("/:id/dashboard", "websites.dashboard", WebsitesPrefix)
)
