package controllers

import (
	"net/http"
	"strconv"

	"github.com/stevenleeg/gobb/models"
	"github.com/stevenleeg/gobb/utils"
)

func AdminBoards(w http.ResponseWriter, r *http.Request) {
	currentUser := utils.GetCurrentUser(r)
	if currentUser == nil || !currentUser.IsAdmin() {
		http.NotFound(w, r)
		return
	}

	db := models.GetDbSession()
	// Creating a board
	if r.Method == "POST" && r.FormValue("create_board") != "" {
		name := r.FormValue("title")
		desc := r.FormValue("description")
		formOrder := r.FormValue("order")
		var order int

		if formOrder != "" {
			if len(formOrder) == 0 {
				order = 1
			} else {
				order, _ = strconv.Atoi(formOrder)
			}
		} else {
			order = 1
		}

		board := models.NewBoard(name, desc, order)

		db.Insert(board)
	}

	// Update the boards
	if r.Method == "POST" && r.FormValue("update_boards") != "" {
		err := r.ParseForm()

		// loop through the post data, entries correspond via index in the map
		for i := 0; i < len(r.Form["board_id"]); i++ {
			// basically repeat the process for inserting a board
			formID, _ := strconv.Atoi(r.Form["board_id"][i])
			id := int64(formID)
			name := r.Form["name"][i]
			desc := r.Form["description"][i]
			formOrder := r.Form["order"][i]
			var order int

			if formOrder != "" {
				if len(formOrder) == 0 {
					order = 1
				} else {
					order, _ = strconv.Atoi(formOrder)
				}
			} else {
				order = 1
			}
			board := models.UpdateBoard(name, desc, order, id)

			db.Update(board)
		}

		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	// Delete a board
	if id := r.FormValue("delete"); id != "" {
		obj, _ := db.Get(&models.Board{}, id)

		if obj == nil {
			http.NotFound(w, r)
			return
		}

		board := obj.(*models.Board)
		board.Delete()
	}

	boards, _ := models.GetBoards()

	utils.RenderTemplate(w, r, "admin_boards.html", map[string]interface{}{
		"boards": boards,
	}, nil)
}
