package main
import (
	forum "forum/Backend"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)
func main() {
	fs := http.FileServer(http.Dir("Style"))
	http.Handle("/Style/", http.StripPrefix("/Style/", fs))
	posts := http.FileServer(http.Dir("Posts"))
	http.Handle("/Posts/", http.StripPrefix("/Posts/", posts))
	profileImages := http.FileServer(http.Dir("ProfileImages"))
	http.Handle("/ProfileImages/", http.StripPrefix("/ProfileImages/", profileImages))
	http.HandleFunc("/", forum.Login)
	http.HandleFunc("/SignUp", forum.SignUpHandler)
	http.HandleFunc("/LogOut", forum.LogoutHandler)
	http.HandleFunc("/HomePage", forum.AuthMiddleware(forum.HomeHandler))
	http.HandleFunc("/CreatePost",forum.AuthMiddleware(forum.CreatePostHandler))
	http.HandleFunc("/Profile", forum.AuthMiddleware(forum.ProfileHandler))
	http.HandleFunc("/CommentHandler", forum.CommentHandler)
	http.HandleFunc("/ProfileImageHandler", forum.ProfileImageHandler)
	http.HandleFunc("/CommentLikeHandle", forum.CommentLikeHandle)
	http.HandleFunc("/googleSignIn", forum.GoogleSignInHandler)
	
	/*http.HandleFunc("/", forum.Login)
	http.HandleFunc("/SignUp", forum.SignUpHandler)
	http.HandleFunc("/HomePage", forum.HomeHandler)
	http.HandleFunc("/CreatePost", forum.CreatePostHandler)
	http.HandleFunc("/Profile", forum.ProfileHandler)*/
	forum.CreateTables()
	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}