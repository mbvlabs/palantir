package routes

import (
	"palantir/internal/routing"
)

var CollectCreate = routing.NewSimpleRoute(
	"/collect",
	"api.collect",
	APIPrefix,
)
