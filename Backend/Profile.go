package forum

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

type ProfileTemplateData struct {
	Posts         []Post
	Username      string
	ProfileImg    string
	LikedPosts    []string
	LikedComments []string
	Createdposts  []string
	ClickedButton string
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	var user Account
	user = GetUserData(w, r)

	filterValue := r.FormValue("filterValue")
	// Check if the Content-Type is application/json
	if r.Header.Get("Content-Type") == "application/json" {

		var request Request
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Failed to decode JSON request", http.StatusBadRequest)
			return
		}
		// Process the request
		switch request.RequestType {
		case "like":
			HandleLikeRequest(w, r, user, request)
		case "delete":
			HandleDeleteRequest(request)
		default:
			http.Error(w, "Invalid request type", http.StatusBadRequest)
			return
		}
	}

	switch filterValue {
	case "CreatedPosts":
		ProfilePost(w, r, "CreatedPosts", user)
	case "LikedPosts":
		ProfilePost(w, r, "LikedPosts", user)
	default:
		ProfilePost(w, r, "CreatedPosts", user)
	}

}

func ProfilePost(w http.ResponseWriter, r *http.Request, filter string, user Account) {
	var posts []Post
	clickedButton := "CreatedPosts"

	//Get Created Posts id
	Createdposts := GetCreatedPosts(user.Id)
	LikedPostsId, Likedposts := GetLikedPosts(user.Id)
	Likedcomments := GetLikedComments(user.Id)

	//To do the filter adjust the fetchPostsFromDB function to accept filters
	if filter == "CreatedPosts" {
		posts, err = fetchPostsFromDB(true, "ProfileFilter", Createdposts)
		if err != nil {
			log.Fatal(err)
		}
		clickedButton = "CreatedPosts"
	} else if filter == "LikedPosts" {
		posts, err = fetchPostsFromDB(true, "ProfileFilter", LikedPostsId)
		if err != nil {
			log.Fatal(err)
		}
		clickedButton = "LikedPosts"
	}

	//Reverse Posts from new to old
	reverseArray(posts)

	data := ProfileTemplateData{
		Posts:         posts,
		Username:      user.Username,
		ProfileImg:    user.ProfileImg,
		LikedPosts:    Likedposts,
		LikedComments: Likedcomments,
		Createdposts:  Createdposts,
		ClickedButton: clickedButton,
	}

	// Reload the templates by parsing the template files
	tmpl, err := template.ParseFiles("./Pages/Profile.html", "./Pages/nav.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
