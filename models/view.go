package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

type View struct {
	ID     string    `db:"id"`
	Post   *Post     `db:"-"`
	PostID int64     `db:"post_id"`
	User   *User     `db:"-"`
	UserID int64     `db:"user_id"`
	Time   time.Time `db:"time"`
}

func AddView(user *User, post *Post) *View {
	db := GetDbSession()

	// Generate the hash of userid and post id
	h := md5.New()
	hash := fmt.Sprintf("%d_%d", user.ID, post.ID)
	h.Write([]byte(hash))
	hash = hex.EncodeToString(h.Sum(nil))

	var view *View
	obj, _ := db.Get(&View{}, hash)
	if obj == nil {
		view = &View{
			ID:     hash,
			Post:   post,
			PostID: post.ID,
			User:   user,
			UserID: user.ID,
			Time:   time.Now(),
		}

		db.Insert(view)
	} else {
		view = obj.(*View)
		view.User = user
		view.Post = post
		view.Time = time.Now()

		db.Update(view)
	}

	return view
}
