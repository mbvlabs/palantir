// Package models contains data models and validation logic.
package models

type (
	user     struct{}
	token    struct{}
	website  struct{}
	pageview struct{}
	event    struct{}
)

var (
	User     user
	Token    token
	Website  website
	Pageview pageview
	Event    event
)
