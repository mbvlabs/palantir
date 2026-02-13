package routes

import (
	"palantir/internal/routing"
)

const APIPrefix = "/api"

var Health = routing.NewSimpleRoute(
	"/health",
	"api.health",
	APIPrefix,
)
