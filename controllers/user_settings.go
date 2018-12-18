package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/stevenleeg/gobb/config"
	"github.com/stevenleeg/gobb/models"
	"github.com/stevenleeg/gobb/utils"
)

func UserSettings(w http.ResponseWriter, r *http.Request) {
	enableSignatures, _ := config.Config.GetBool("gobb", "enable_signatures")

	userID, _ := strconv.Atoi(mux.Vars(r)["id"])
	currentUser := utils.GetCurrentUser(r)

	if currentUser == nil || int64(userID) != currentUser.ID {
		http.NotFound(w, r)
		return
	}

	success := false
	var formError string
	if r.Method == "POST" {
		db := models.GetDbSession()
		currentUser.Avatar = r.FormValue("avatar_url")
		currentUser.UserTitle = r.FormValue("user_title")
		currentUser.StylesheetURL = sql.NullString{
			Valid:  true,
			String: r.FormValue("stylesheet_url"),
		}
		if r.FormValue("signature") == "" {
			currentUser.Signature = sql.NullString{
				Valid:  false,
				String: r.FormValue("signature"),
			}
		} else {
			currentUser.Signature = sql.NullString{
				Valid:  true,
				String: r.FormValue("signature"),
			}
		}

		// Change hiding settings
		currentUser.HideOnline = false
		if r.FormValue("hide_online") == "1" {
			currentUser.HideOnline = true
		}

		// Update password?
		old_pass := r.FormValue("password_old")
		new_pass := r.FormValue("password_new")
		new_pass2 := r.FormValue("password_new2")
		if old_pass != "" {
			user, err := models.AuthenticateUser(currentUser.Username, old_pass)
			if user == nil || err != nil {
				formError = "Invalid password"
			} else if len(new_pass) < 5 {
				formError = "Password must be greater than 4 characters"
			} else if new_pass != new_pass2 {
				formError = "Passwords didn't match"
			} else {
				currentUser.SetPassword(new_pass)
				session, _ := utils.GetCookieStore(r).Get(r, "sirsid")
				session.Values["password"] = new_pass
				session.Save(r, w)
			}
		}

		if formError == "" {
			db.Update(currentUser)
			success = true
		}
	}

	stylesheet := ""
	if currentUser.StylesheetURL.Valid {
		stylesheet = currentUser.StylesheetURL.String
	}
	signature := ""
	if currentUser.Signature.Valid {
		signature = currentUser.Signature.String
	}

	utils.RenderTemplate(w, r, "user_settings.html", map[string]interface{}{
		"error":             formError,
		"success":           success,
		"user_stylesheet":   stylesheet,
		"user_signature":    signature,
		"enable_signatures": enableSignatures,
	}, nil)
}
