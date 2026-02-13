package routes

import (
	"palantir/internal/routing"
)

var HomePage = routing.NewSimpleRoute(
	"/",
	"pages.home",
	"",
)
