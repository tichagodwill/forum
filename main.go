package main

import (
    "forum/Backend" // Assuming your forum package is in a directory named "Backend"
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

    // Registering route handlers
    http.HandleFunc("/SignUp", forum.SignUpHandler)
    http.HandleFunc("/LogOut", forum.LogoutHandler)
    http.HandleFunc("/HomePage", forum.AuthMiddleware(forum.HomeHandler))
    http.HandleFunc("/CreatePost", forum.AuthMiddleware(forum.CreatePostHandler))
    http.HandleFunc("/Profile", forum.AuthMiddleware(forum.ProfileHandler))
    http.HandleFunc("/CommentHandler", forum.CommentHandler)
    http.HandleFunc("/ProfileImageHandler", forum.ProfileImageHandler)
    http.HandleFunc("/CommentLikeHandle", forum.CommentLikeHandle)
    http.HandleFunc("/googleSignIn", forum.GoogleSignInHandler)
    // Registering error handlers
    http.HandleFunc("/HandleBadRequest", forum.HandleBadRequest)
    http.HandleFunc("/HandleNotFound", forum.HandleNotFound)
    http.HandleFunc("/HandleInternalServerError", forum.HandleInternalServerError)

    // Catch-all route handler for undefined routes
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            forum.HandleNotFound(w, r)
            return
        }
        forum.Login(w, r)
    })

    // Create tables if needed
    forum.CreateTables()

    log.Println("Server started on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
