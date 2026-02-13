package routes

import (
	"palantir/internal/routing"
)

const TrackingPrefix = "/t"

var TrackingScript = routing.NewSimpleRoute(
	"/script.js",
	"tracking.script",
	TrackingPrefix,
)
