package utils

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/stevenleeg/gobb/config"
	"github.com/stevenleeg/gobb/models"
)

var Store *sessions.CookieStore

func GetCookieStore(r *http.Request) *sessions.CookieStore {
	if Store == nil {
		cookieKey, _ := config.Config.GetString("gobb", "cookie_key")
		Store = sessions.NewCookieStore([]byte(cookieKey))
	}

	return Store
}

func GetCurrentUser(r *http.Request) *models.User {
	cached := context.Get(r, "user")
	if cached != nil {
		return cached.(*models.User)
	}

	session, _ := GetCookieStore(r).Get(r, "sirsid")
	sessionUsername := session.Values["username"]
	sessionPassword := session.Values["password"]

	if sessionUsername == nil || sessionPassword == nil {
		return nil
	}

	currentUser, err := models.AuthenticateUser(sessionUsername.(string), sessionPassword.(string))
	if err != nil {
		return nil
	}

	context.Set(r, "user", currentUser)
	return currentUser
}
