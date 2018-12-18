package models

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/stevenleeg/gobb/config"
)

type User struct {
	ID            int64          `db:"id"`
	GroupID       int64          `db:"group_id"`
	CreatedOn     time.Time      `db:"created_on"`
	Username      string         `db:"username"`
	Password      string         `db:"password"`
	Avatar        string         `db:"avatar"`
	Signature     sql.NullString `db:"signature"`
	Salt          string         `db:"salt"`
	StylesheetURL sql.NullString `db:"stylesheet_url"`
	UserTitle     string         `db:"user_title"`
	LastSeen      time.Time      `db:"last_seen"`
	HideOnline    bool           `db:"hide_online"`
	LastUnreadAll pq.NullTime    `db:"last_unread_all"`
}

func NewUser(username, password string) *User {
	user := &User{
		CreatedOn: time.Now(),
		Username:  username,
		LastSeen:  time.Now(),
	}

	user.SetPassword(password)
	return user
}

func AuthenticateUser(username, password string) (*User, error) {
	db := GetDbSession()
	user := &User{}
	err := db.SelectOne(user, "SELECT * FROM users WHERE username=$1", username)
	if err != nil {
		log.Printf("[error] Cannot select user '%s', (%s)\n", username, err.Error())
		return nil, err
	}

	if user.ID == 0 {
		return nil, errors.New("Inval username/password")
	}

	hasher := sha1.New()
	io.WriteString(hasher, password)
	io.WriteString(hasher, user.Salt)
	password = base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	if password != user.Password {
		return nil, errors.New("Inval username/password")
	}

	// Update the user's last seen
	user.LastSeen = time.Now()
	db.Update(user)

	return user, nil
}

func GetUserCount() (int64, error) {
	db := GetDbSession()

	count, err := db.SelectInt("SELECT COUNT(*) FROM users")
	if err != nil {
		log.Printf("[error] Error selecting user count (%s)\n", err.Error())
		return 0, errors.New("Database error: " + err.Error())
	}

	return count, nil
}

func GetLatestUser() (*User, error) {
	db := GetDbSession()

	user := &User{}
	err := db.SelectOne(user, "SELECT * FROM users ORDER BY created_on DESC LIMIT 1")

	if err != nil {
		log.Printf("[error] Error selecting latest user (%s)\n", err.Error())
		return nil, fmt.Errorf("Database error: %s", err.Error())
	}

	if user.Username == "" {
		return nil, nil
	}

	return user, nil
}

func GetOnlineUsers() (users []*User) {
	db := GetDbSession()
	db.Select(&users, "SELECT * FROM users WHERE last_seen > current_timestamp - interval '5 minutes' AND he_online != true")

	return users
}

func GetUser(ID int) (*User, error) {
	db := GetDbSession()
	obj, err := db.Get(&User{}, ID)
	if obj == nil {
		return nil, err
	}

	return obj.(*User), err
}

// Converts the given string into an appropriate hash, resets the salt,
// and sets the Password attribute. Does *not* commit to the database.
func (user *User) SetPassword(password string) {
	var int_salt int32
	binary.Read(rand.Reader, binary.LittleEndian, &int_salt)
	salt := strconv.Itoa(int(int_salt))

	hasher := sha1.New()
	io.WriteString(hasher, password)
	io.WriteString(hasher, salt)
	user.Password = base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	user.Salt = salt
}

func (user *User) IsAdmin() bool {
	if user.GroupID == 2 {
		return true
	}

	return false
}

func (user *User) CanModerate() bool {
	if user.GroupID > 0 {
		return true
	}

	return false
}

func (user *User) GetPostCount() int64 {
	db := GetDbSession()
	count, err := db.SelectInt("SELECT COUNT(*) FROM posts WHERE author_id=$1", user.ID)

	if err != nil {
		return 0
	}

	return count
}

func (user *User) GetPosts(page int) []*Post {
	db := GetDbSession()
	var posts []*Post

	postsPerPage, _ := config.Config.GetInt64("gobb", "posts_per_page")
	offset := postsPerPage * int64(page)

	_, err := db.Select(&posts, "SELECT * FROM posts WHERE author_id=$1 ORDER BY created_on DESC LIMIT $2 OFFSET $3", user.ID, postsPerPage, offset)

	if err != nil {
		log.Printf("[error] Could not get user's posts (%s)", err.Error())
	}

	return posts
}
