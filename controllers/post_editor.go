package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/stevenleeg/gobb/models"
	"github.com/stevenleeg/gobb/utils"
)

func renderPostEditor(
	w http.ResponseWriter,
	r *http.Request,
	board *models.Board,
	post *models.Post,
	err error) {

	utils.RenderTemplate(w, r, "post_editor.html", map[string]interface{}{
		"board": board,
		"post":  post,
		"error": err,
	}, map[string]interface{}{
		"ShowTitleField": func() bool {
			if post == nil {
				return true
			}

			return !post.ParentID.Valid
		},
	})
}

func PostEditor(w http.ResponseWriter, r *http.Request) {
	db := models.GetDbSession()

	var err error
	var board *models.Board
	var post *models.Post

	// Attempt to get a board
	board_id_str := mux.Vars(r)["board_id"]
	if board_id_str != "" {
		board_id, _ := strconv.Atoi(board_id_str)
		board, err = models.GetBoard(board_id)
	}

	// Otherwise, a post
	post_id_str := r.FormValue("post_id")
	if post_id_str != "" {
		post_id, _ := strconv.Atoi(post_id_str)
		post_tmp, _ := db.Get(&models.Post{}, post_id)
		post = post_tmp.(*models.Post)
	}

	if err != nil {
		fmt.Println("something went wrong")
		http.NotFound(w, r)
		return
	}

	currentUser := utils.GetCurrentUser(r)
	if currentUser == nil {
		http.NotFound(w, r)
		return
	}

	if post != nil && (post.AuthorID != currentUser.ID && !currentUser.CanModerate()) {
		http.NotFound(w, r)
		return
	}

	if r.Method == "POST" {
		title := r.FormValue("title")
		content := r.FormValue("content")

		if post == nil {
			post = models.NewPost(currentUser, board, title, content)
			post.LatestReply = time.Now()

			err = post.Validate()
			if err != nil {
				renderPostEditor(w, r, board, post, err)
				return
			}

			err = db.Insert(post)
		} else {
			post.Title = title
			post.Content = content
			post.LastEdit = time.Now()
			post.LatestReply = time.Now()

			err = post.Validate()
			if err != nil {
				renderPostEditor(w, r, board, post, err)
				return
			}

			_, err = db.Update(post)
		}

		if err != nil {
			fmt.Printf("[error] Could not save post (%s)", err.Error())
			return
		}

		http.Redirect(w, r, post.GetLink(), http.StatusFound)
		return
	}

	renderPostEditor(w, r, board, post, err)
}
