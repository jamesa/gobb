package controllers

import (
	"fmt"
	"net/http"

	"github.com/stevenleeg/gobb/models"
	"github.com/stevenleeg/gobb/utils"
)

func Login(w http.ResponseWriter, r *http.Request) {
	if utils.GetCurrentUser(r) != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var error string
		_, err := models.AuthenticateUser(username, password)
		if err != nil {
			error = "Invalid username or password"
		}

		if error != "" {
			utils.RenderTemplate(w, r, "login.html", map[string]interface{}{
				"error": error,
			}, nil)
			return
		}

		session, _ := utils.GetCookieStore(r).Get(r, "sirsid")
		session.Values["username"] = username
		session.Values["password"] = password

		err = session.Save(r, w)
		if err != nil {
			fmt.Printf("[error] Could not save session (%s)\n", err.Error())
		}

		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	utils.RenderTemplate(w, r, "login.html", nil, nil)
}
