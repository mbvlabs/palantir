package cookies

import (
	"context"
	"strings"

	"palantir/config"
	"palantir/internal/renderer"
	"github.com/google/uuid"
	"palantir/models"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v5"
)

var AppKey renderer.CookieKey = "app_cookie_context"
const ReturnToKey = "return_to"

const (
	isAuthenticated = "is_authenticated"
	isAdmin = "is_admin"
	userID = "user_id"
)

type App struct {
	UserID uuid.UUID
	IsAdmin bool
	IsAuthenticated bool
}
func CreateAppSession(c *echo.Context, user models.User) error {
	sess, err := session.Get(config.AppCookieSessionName, c)
	if err != nil {
		return err
	}

	sess.Values[isAuthenticated] = true
	sess.Values[isAdmin] = user.IsAdmin
	sess.Values[userID] = user.ID.String()

	return sess.Save(c.Request(), c.Response())
}

func DestroyAppSession(c *echo.Context) error {
	sess, err := session.Get(config.AppCookieSessionName, c)
	if err != nil {
		return err
	}

	sess.Options.MaxAge = -1
	return sess.Save(c.Request(), c.Response())
}

func GetAppCtx(ctx context.Context) App {
	appCtx, ok := ctx.Value(AppKey).(App)
	if !ok {
		return App{}
	}

	return appCtx
}

func GetApp(c *echo.Context) App {
	sess, err := session.Get(config.AppCookieSessionName, c)
	if err != nil {
		return App{}
	}

	app := App{}

	if v, ok := sess.Values[isAuthenticated].(bool); ok {
		app.IsAuthenticated = v
	}
	if v, ok := sess.Values[isAdmin].(bool); ok {
		app.IsAdmin = v
	}
	if v, ok := sess.Values[userID].(string); ok {
		app.UserID, _ = uuid.Parse(v)
	}

	return app
}

func SetReturnTo(c *echo.Context, returnTo string) error {
	sess, err := session.Get(config.AppCookieSessionName, c)
	if err != nil {
		return err
	}

	if returnTo == "" {
		delete(sess.Values, ReturnToKey)
	} else {
		sess.Values[ReturnToKey] = returnTo
	}

	return sess.Save(c.Request(), c.Response())
}

func GetReturnTo(c *echo.Context) string {
	sess, err := session.Get(config.AppCookieSessionName, c)
	if err != nil {
		return ""
	}

	returnTo, _ := sess.Values[ReturnToKey].(string)
	return strings.TrimSpace(returnTo)
}

func ResolveReturnTo(c *echo.Context, fallback string) string {
	returnTo := GetReturnTo(c)
	if !isSafeReturnTo(returnTo) {
		return fallback
	}

	return returnTo
}

func isSafeReturnTo(returnTo string) bool {
	if returnTo == "" {
		return false
	}

	return strings.HasPrefix(returnTo, "/") &&
		!strings.HasPrefix(returnTo, "//")
}
