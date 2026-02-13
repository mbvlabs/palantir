package routes

import (
	"palantir/internal/routing"
)

const WebsitesPrefix = "/websites"

var WebsiteIndex = routing.NewSimpleRoute(
	"",
	"websites.index",
	WebsitesPrefix,
)

var WebsiteNew = routing.NewSimpleRoute(
	"/new",
	"websites.new",
	WebsitesPrefix,
)

var WebsiteCreate = routing.NewSimpleRoute(
	"",
	"websites.create",
	WebsitesPrefix,
)

var WebsiteShow = routing.NewRouteWithUUIDID(
	"/:id",
	"websites.show",
	WebsitesPrefix,
)

var WebsiteEdit = routing.NewRouteWithUUIDID(
	"/:id/edit",
	"websites.edit",
	WebsitesPrefix,
)

var WebsiteUpdate = routing.NewRouteWithUUIDID(
	"/:id",
	"websites.update",
	WebsitesPrefix,
)

var WebsiteDestroy = routing.NewRouteWithUUIDID(
	"/:id",
	"websites.destroy",
	WebsitesPrefix,
)

var WebsiteDashboard = routing.NewRouteWithUUIDID(
	"/:id/dashboard",
	"websites.dashboard",
	WebsitesPrefix,
)

var WebsiteDashboardLive = routing.NewRouteWithUUIDID(
	"/:id/dashboard/live",
	"websites.dashboard.live",
	WebsitesPrefix,
)
