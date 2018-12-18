package models

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/stevenleeg/gobb/config"
)

type Post struct {
	ID          int64         `db:"id"`
	BoardID     int64         `db:"board_id"`
	ParentID    sql.NullInt64 `db:"parent_id"`
	Author      *User         `db:"-"`
	AuthorID    int64         `db:"author_id"`
	Title       string        `db:"title"`
	Content     string        `db:"content"`
	CreatedOn   time.Time     `db:"created_on"`
	LatestReply time.Time     `db:"latest_reply"`
	LastEdit    time.Time     `db:"last_edit"`
	Sticky      bool          `db:"sticky"`
	Locked      bool          `db:"locked"`
}

// Initializes a new struct, adds some data, and returns the pointer to it
func NewPost(author *User, board *Board, title, content string) *Post {
	post := &Post{
		BoardID:   board.ID,
		AuthorID:  author.ID,
		Title:     title,
		Content:   content,
		CreatedOn: time.Now(),
		Sticky:    false,
	}

	return post
}

func GetPost(ID int) (*Post, error) {
	db := GetDbSession()
	obj, err := db.Get(&Post{}, ID)
	if obj == nil {
		return nil, err
	}

	return obj.(*Post), err
}

// Returns a pointer to the OP and a slice of post pointers for the given
// page number in the thread.
func GetThread(parentID, pageID int) (error, *Post, []*Post) {
	db := GetDbSession()

	op, err := db.Get(Post{}, parentID)
	if err != nil || op == nil {
		fmt.Printf("Something weird is going on here: parentID: %d, pageID: %d", parentID, pageID)
		return fmt.Errorf("[error] Could not get parent (%d)", parentID), nil, nil
	}

	postsPerPage, err := config.Config.GetInt64("gobb", "posts_per_page")
	if err != nil {
		postsPerPage = 15
	}

	i_begin := (int64(pageID) * (postsPerPage)) - 1
	// The first page already has the OP, which isn't included
	if pageID == 0 {
		postsPerPage--
		i_begin++
	}

	var childPosts []*Post
	db.Select(&childPosts, "SELECT * FROM posts WHERE parent_id=$1 ORDER BY created_on ASC LIMIT $2 OFFSET $3", parentID, postsPerPage, i_begin)

	return nil, op.(*Post), childPosts
}

// Returns the number of posts (on every board/thread)
func GetPostCount() (int64, error) {
	db := GetDbSession()

	count, err := db.SelectInt("SELECT COUNT(*) FROM posts")
	if err != nil {
		fmt.Printf("[error] Error selecting post count (%s)\n", err.Error())
		return 0, errors.New("Database error: " + err.Error())
	}

	return count, nil
}

// Post-SELECT hook for gorp which adds a pointer to the author
// to the Post's struct
func (post *Post) PostGet(s gorp.SqlExecutor) error {
	db := GetDbSession()
	user, _ := db.Get(User{}, post.AuthorID)

	if user == nil {
		return errors.New("Could not find post's author")
	}

	post.Author = user.(*User)

	return nil
}

// Ensures that a post is valid
func (post *Post) Validate() error {
	if post.BoardID == 0 {
		return errors.New("Board does not exist")
	}

	if len(post.Content) <= 3 {
		return errors.New("Post must be longer than three characters")
	}

	if !post.ParentID.Valid && len(post.Title) <= 3 {
		return errors.New("Post title must be longer than three characters")
	}

	return nil
}

// This is used primarily for threads. It will find the latest
// post in a thread, allowing for things like "last post was 10
// minutes ago.
func (post *Post) GetLatestPost() *Post {
	db := GetDbSession()
	latest := &Post{}

	db.SelectOne(latest, "SELECT * FROM posts WHERE parent_id=$1 ORDER BY created_on DESC LIMIT 1", post.ID)

	return latest
}

// Returns the number of pages contained by a thread. This won't work on
// post structs that have ParentIds.
func (post *Post) GetPagesInThread() int {
	db := GetDbSession()
	count, err := db.SelectInt("SELECT COUNT(*) FROM posts WHERE parent_id=$1", post.ID)

	if err != nil {
		fmt.Printf("[error] Could not get post count (%s)\n", err.Error())
	}

	postsPerPage, err := config.Config.GetInt64("gobb", "posts_per_page")

	if err != nil {
		postsPerPage = 15
	}

	if count == postsPerPage {
		return 1
	}

	return int(math.Floor(float64(count) / float64(postsPerPage)))
}

// This function tells us which page this particular post is in
// within a thread based on the current value of posts_per_page
func (post *Post) GetPageInThread() int {
	postsPerPage, err := config.Config.GetInt64("gobb", "posts_per_page")
	if err != nil {
		postsPerPage = 15
	}

	db := GetDbSession()
	n, err := db.SelectInt(`
        WITH thread AS (
                SELECT posts.*,
                ROW_NUMBER() OVER(ORDER BY posts.id) AS position
                FROM posts WHERE parent_id=$1)
        SELECT 
            posts.position
        FROM 
            thread posts
        WHERE 
            posts.id=$2 AND 
            posts.parent_id=$1;
    `, post.ParentID, post.ID)

	return int(math.Floor(float64(n) / float64(postsPerPage)))
}

// Used when deleting a thread. This deletes all posts who are
// children of the OP.
func (post *Post) DeleteAllChildren() error {
	db := GetDbSession()

	_, err := db.Exec("DELETE FROM posts WHERE parent_id=$1", post.ID)
	return err
}

// Get the thread id for a post
func (post *Post) GetThreadID() int64 {
	if post.ParentID.Valid {
		return post.ParentID.Int64
	} else {
		return post.ID
	}
}

// Generate a link to a post
func (post *Post) GetLink() string {
	return fmt.Sprintf("/board/%d/%d?page=%d#post_%d", post.BoardID, post.GetThreadID(), post.GetPageInThread(), post.ID)
}
