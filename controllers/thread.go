package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/stevenleeg/gobb/config"
	"github.com/stevenleeg/gobb/models"
	"github.com/stevenleeg/gobb/utils"
)

func Thread(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		pageID = 0
	}

	boardID, _ := strconv.Atoi(mux.Vars(r)["board_id"])
	board, err := models.GetBoard(boardID)

	postID, err := strconv.Atoi(mux.Vars(r)["post_id"])
	err, op, posts := models.GetThread(postID, pageID)

	var postingError error

	currentUser := utils.GetCurrentUser(r)
	if r.Method == "POST" {
		db := models.GetDbSession()
		title := r.FormValue("title")
		content := r.FormValue("content")

		if currentUser == nil {
			http.NotFound(w, r)
			return
		}

		if op.Locked && !currentUser.CanModerate() {
			http.NotFound(w, r)
			return
		}

		post := models.NewPost(currentUser, board, title, content)
		post.ParentID = sql.NullInt64{int64(postID), true}
		op.LatestReply = time.Now()

		postingError = post.Validate()

		if postingError == nil {
			db.Insert(post)
			db.Update(op)

			if page := post.GetPageInThread(); page != pageID {
				http.Redirect(w, r, fmt.Sprintf("/board/%d/%d?page=%d#post_%d", post.BoardID, op.ID, page, post.ID), http.StatusFound)
			}

			err, op, posts = models.GetThread(postID, pageID)
		}
	}

	if err != nil {
		http.NotFound(w, r)
		fmt.Printf("[error] Something went wrong in posts (%s)\n", err.Error())
		return
	}

	numPages := op.GetPagesInThread()

	if pageID > numPages {
		http.NotFound(w, r)
		return
	}

	var previousText string
	if postingError != nil {
		previousText = r.FormValue("content")
	}

	// Mark the thread as read
	if currentUser != nil {
		models.AddView(currentUser, op)
	}

	utils.RenderTemplate(w, r, "thread.html", map[string]interface{}{
		"board":        board,
		"op":           op,
		"posts":        posts,
		"first_page":   (pageID > 0),
		"prev_page":    (pageID > 1),
		"next_page":    (pageID < numPages-1),
		"last_page":    (pageID < numPages),
		"pageID":       pageID,
		"postingError": postingError,
		"previousText": previousText,
	}, map[string]interface{}{

		"CurrentUserCanModerateThread": func(thread *models.Post) bool {
			currentUser := utils.GetCurrentUser(r)
			if currentUser == nil {
				return false
			}

			return (currentUser.CanModerate() && thread.ParentID.Valid == false)
		},

		"CurrentUserCanDeletePost": func(thread *models.Post) bool {
			currentUser := utils.GetCurrentUser(r)
			if currentUser == nil {
				return false
			}

			return (currentUser.ID == thread.AuthorID) || currentUser.CanModerate()
		},

		"CurrentUserCanEditPost": func(post *models.Post) bool {
			currentUser := utils.GetCurrentUser(r)
			if currentUser == nil {
				return false
			}

			return (currentUser.ID == post.AuthorID || currentUser.CanModerate())
		},

		"CurrentUserCanModerate": func() bool {
			currentUser := utils.GetCurrentUser(r)
			if currentUser == nil {
				return false
			}

			return currentUser.CanModerate()
		},

		"SignaturesEnabled": func() bool {
			enableSignatures, _ := config.Config.GetBool("gobb", "enable_signatures")
			return enableSignatures
		},

		"CurrentUserCanReply": func(post *models.Post) bool {
			currentUser := utils.GetCurrentUser(r)
			if currentUser != nil && (!post.Locked || currentUser.CanModerate()) {
				return true
			}
			return false
		},
	})
}
