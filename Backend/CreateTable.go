package forum

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func CreateTables() {
	//Accounts db
	//Create Accounts Table
	Accountsdb, err = sql.Open("sqlite3", "./DataBase/accounts.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create the account table if it doesn't exist
	AccountsTableQuery := `CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		Email TEXT,
		Username TEXT UNIQUE,
		Password TEXT,
		SessionID TEXT,
		Expiration TIMESTAMP,
		UserImg TEXT DEFAULT 'ProfileImage.png'
	);`

	_, err = Accountsdb.Exec(AccountsTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	//Posts db
	Postsdb, err = sql.Open("sqlite3", "./DataBase/Posts.db")
	if err != nil {
		log.Fatal(err)
	}

	//Create Table to store the post
	PostsTableQuery := `CREATE TABLE IF NOT EXISTS posts (
		post_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		username TEXT,
		userImg TEXT,
		title TEXT,
		content TEXT,
		image TEXT,
		category TEXT,
		like INTEGER,
		dislike INTEGER,
		timestamp TEXT DEFAULT (strftime('%d/%m/%Y', 'now', 'localtime')),
		FOREIGN KEY (user_id) REFERENCES accounts(id));`

	_, err = Postsdb.Exec(PostsTableQuery)
	if err != nil {
		log.Fatal(err)
	}


	//Liked Posts
	LikedPostsdb, err = sql.Open("sqlite3", "./DataBase/Posts.db")
	if err != nil {
		log.Fatal(err)
	}

	//Create Table to store the post
	LikedPostsTableQuery := `CREATE TABLE IF NOT EXISTS LikedPosts (
		user_id INTEGER,
		post_id INTEGER,
		type    Text,
		FOREIGN KEY (user_id) REFERENCES accounts(id));`

	_, err = Postsdb.Exec(LikedPostsTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	//Liked Posts
	Commentsdb, err = sql.Open("sqlite3", "./DataBase/Comments.db")
	if err != nil {
		log.Fatal(err)
	}

	//Create Table to store the post
	CommentsTableQuery := `CREATE TABLE IF NOT EXISTS comments (
		comment_id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		userName Text,
		post_id INTEGER,
		text    Text,
		ProfileImage TEXT,
		CLike INTEGER DEFAULT 0,
		CDislike INTEGER DEFAULT 0,
		timestamp TEXT DEFAULT (strftime('%d/%m/%Y', 'now', 'localtime')));`

	_, err = Commentsdb.Exec(CommentsTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	LCommentsTableQuery := `CREATE TABLE IF NOT EXISTS Likedcomments (
		comment_id INTEGER,
		user_id INTEGER,
		Type TEXT);`

	_, err = Commentsdb.Exec(LCommentsTableQuery)
	if err != nil {
		log.Fatal(err)
	}

}
