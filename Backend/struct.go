package forum

import (
	"database/sql"
	
)

type Account struct {
	Id          int
	Email       string
	Username    string
	Password    string
	ProfileImg  string
}

var Accountsdb *sql.DB
//var Sessionsdb *sql.DB
var Postsdb *sql.DB
var LikedPostsdb *sql.DB
var Commentsdb *sql.DB
var err error
var guest bool

var Posts []Post
type Post struct {
	PostId    int
	UserId    int
	UserName  string
	UserImg   string
	Title     string
	Content   string
	Image     string
	Category  []string
	Time      string
	Like      int
	Dislike   int
	Comments []Comment
}

type Comment struct {
	//PostId    int
	CommentId int
	UserName    string
	Text    string
	Time    string
	CLike int
	CDislike  int
	ProfileImage string
}

type Request struct {
	RequestType string `json:"RequestType"`
	//For Like Request
	Type    string `json:"Type"`
	ID      string `json:"ID"`
	Checked bool   `json:"Checked"`
	//For Filter
	Categories []string `json:"Categories"`
}