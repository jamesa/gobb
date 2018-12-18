package controllers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/stevenleeg/gobb/models"
	"github.com/stevenleeg/gobb/utils"
)

func User(w http.ResponseWriter, r *http.Request) {
	db := models.GetDbSession()

	userID, err := strconv.Atoi(mux.Vars(r)["id"])

	if err != nil {
		http.NotFound(w, r)
		return
	}

	user, err := db.Get(&models.User{}, userID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	utils.RenderTemplate(w, r, "user.html", map[string]interface{}{
		"user": user,
	}, nil)
}
