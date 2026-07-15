// Package controllers provides HTTP handlers for the web application.
package controllers

import (
	"go.uber.org/fx"
	"palantir/controllers/api"
	"palantir/router"
)

var otherCache = NewCacheBuilder[string]().WithSize(2).Build

var constructors = fx.Provide(
	otherCache,
	NewPages,
	NewAssets,
	api.NewAPI,
	NewSessions,
	NewResetPasswords,
	NewWebsites,
	NewDashboard,
	NewCollect,
	NewTracking,
)

var Module = fx.Module(
	"controllers",
	constructors,
	fx.Invoke(func(r *router.Router, c Pages) error {
		return c.RegisterRoutes(r)
	}),
	fx.Invoke(func(r *router.Router, c Assets) error {
		return c.RegisterRoutes(r)
	}),
	fx.Invoke(func(r *router.Router, c api.API) error {
		return c.RegisterRoutes(r)
	}),
	fx.Invoke(func(r *router.Router, c Sessions) error {
		return c.RegisterRoutes(r)
	}),
	fx.Invoke(func(r *router.Router, c ResetPasswords) error {
		return c.RegisterRoutes(r)
	}),
	fx.Invoke(func(r *router.Router, c Websites) error {
		return c.RegisterRoutes(r)
	}),
	fx.Invoke(func(r *router.Router, c Dashboard) error {
		return c.RegisterRoutes(r)
	}),
	fx.Invoke(func(r *router.Router, c Collect) error {
		return c.RegisterRoutes(r)
	}),
	fx.Invoke(func(r *router.Router, c Tracking) error {
		return c.RegisterRoutes(r)
	}),
)
