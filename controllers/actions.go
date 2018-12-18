package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/stevenleeg/gobb/models"
	"github.com/stevenleeg/gobb/utils"
)

func ActionMarkAllRead(w http.ResponseWriter, r *http.Request) {
	user := utils.GetCurrentUser(r)
	if user == nil {
		http.NotFound(w, r)
		return
	}

	db := models.GetDbSession()
	user.LastUnreadAll = pq.NullTime{Time: time.Now(), Valid: true}
	db.Update(user)

	http.Redirect(w, r, "/", http.StatusFound)
}

func ActionStickThread(w http.ResponseWriter, r *http.Request) {
	user := utils.GetCurrentUser(r)
	if !user.CanModerate() {
		http.NotFound(w, r)
		return
	}

	threadID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	db := models.GetDbSession()
	obj, err := db.Get(&models.Post{}, threadID)
	thread := obj.(*models.Post)

	if thread == nil || err != nil {
		http.NotFound(w, r)
		return
	}

	thread.Sticky = !(thread.Sticky)
	db.Update(thread)

	http.Redirect(w, r, fmt.Sprintf("/board/%d/%d", thread.BoardID, thread.ID), http.StatusFound)
}

func ActionLockThread(w http.ResponseWriter, r *http.Request) {
	user := utils.GetCurrentUser(r)
	if !user.CanModerate() {
		http.NotFound(w, r)
		return
	}

	threadID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	db := models.GetDbSession()
	obj, err := db.Get(&models.Post{}, threadID)
	thread := obj.(*models.Post)

	if thread == nil || err != nil {
		http.NotFound(w, r)
		return
	}

	thread.Locked = !(thread.Locked)
	db.Update(thread)

	http.Redirect(w, r, fmt.Sprintf("/board/%d/%d", thread.BoardID, thread.ID), http.StatusFound)
}

func ActionDeleteThread(w http.ResponseWriter, r *http.Request) {
	user := utils.GetCurrentUser(r)

	threadID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	db := models.GetDbSession()
	obj, err := db.Get(&models.Post{}, threadID)
	thread := obj.(*models.Post)

	if thread == nil || err != nil {
		http.NotFound(w, r)
		return
	}

	if (thread.AuthorID != user.ID) && !user.CanModerate() {
		http.NotFound(w, r)
		return
	}

	redirectBoard := true
	if thread.ParentID.Valid {
		redirectBoard = false
	}

	thread.DeleteAllChildren()
	db.Delete(thread)

	if redirectBoard {
		http.Redirect(w, r, fmt.Sprintf("/board/%d", thread.BoardID), http.StatusFound)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/board/%d/%d", thread.BoardID, thread.ParentID.Int64), http.StatusFound)
	}
}

func ActionMoveThread(w http.ResponseWriter, r *http.Request) {
	currentUser := utils.GetCurrentUser(r)
	if currentUser == nil || !currentUser.CanModerate() {
		http.NotFound(w, r)
		return
	}

	threadID, err := strconv.Atoi(r.FormValue("post_id"))
	boardID, err := strconv.Atoi(r.FormValue("to"))

	op, err := models.GetPost(threadID)
	boards, _ := models.GetBoards()

	if op == nil || err != nil {
		http.NotFound(w, r)
		return
	}

	if r.FormValue("to") != "" {
		db := models.GetDbSession()
		targetBoard, _ := models.GetBoard(boardID)
		if targetBoard == nil {
			http.NotFound(w, r)
			return
		}

		_, err := db.Exec("UPDATE posts SET board_id=$1 WHERE parent_id=$2", targetBoard.ID, op.ID)
		op.BoardID = targetBoard.ID
		db.Update(op)
		if err != nil {
			http.NotFound(w, r)
			fmt.Printf("Error moving post: %s\n", err.Error())
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/board/%d/%d", op.BoardID, op.ID), http.StatusFound)
	}

	board, err := models.GetBoard(int(op.BoardID))

	utils.RenderTemplate(w, r, "action_move_thread.html", map[string]interface{}{
		"board":  board,
		"thread": op,
		"boards": boards,
	}, nil)
}
