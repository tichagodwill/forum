package forum

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type HomeTemplateData struct {
	Posts         []Post
	Username      string
	ProfileImg    string
	LikedPosts    []string
	LikedComments []string
	Createdposts  []string
}

var user Account

var Likedposts []string
var Createdposts []string
var Likedcomments []string

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	// Query the user data based on the ID
	if !guest {
		user = GetUserData(w, r)
	}
	if r.Method == "POST" {
		var request Request

		// Check if the Content-Type is application/json
		if r.Header.Get("Content-Type") == "application/json" {
			err := json.NewDecoder(r.Body).Decode(&request)
			if err != nil {
				// Handle other decoding errors
				http.Error(w, "Failed to decode JSON request", http.StatusBadRequest)
				return
			}
			switch request.RequestType {
			case "like":
				HandleLikeRequest(w, r, user, request)
			case "delete":
				HandleDeleteRequest(request)
			default:
				http.Error(w, "Invalid request type", http.StatusBadRequest)
				return
			}

		} else {
			// Handle form data
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Failed to parse form data", http.StatusBadRequest)
				return
			}

			categories := r.Form["categories"]
			ConstructPage(guest, user.Id, w, r, categories)
		}

	} else {

		var All []string
		ConstructPage(guest, user.Id, w, r, All)
	}
}

func ConstructPage(guest bool, Id int, w http.ResponseWriter, r *http.Request, categories []string) {
	Posts = nil
	if len(categories) == 0 {
		Posts, err = fetchPostsFromDB(false, "", categories)
	} else {
		Posts, err = fetchPostsFromDB(true, "Categories", categories)
	}

	if err != nil {
		log.Fatal(err)
	}

	//Reverse Posts from new to old
	reverseArray(Posts)
	if !guest {
		_, Likedposts = GetLikedPosts(Id)
		Createdposts = GetCreatedPosts(Id)
		Likedcomments = GetLikedComments(Id)
	}

	//data

	//fmt.Println("This is LikedPosts:",Likedposts)

	data := HomeTemplateData{
		Posts:         Posts,
		Username:      user.Username,
		ProfileImg:    user.ProfileImg,
		LikedPosts:    Likedposts,
		LikedComments: Likedcomments,
		Createdposts:  Createdposts,
	}

	// Render the template
	var tmpl = template.Must(template.ParseFiles("./Pages/HomePage.html", "./Pages/nav.html"))
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func fetchPostsFromDB(Filtered bool, Type string, Filter []string) ([]Post, error) {
	Posts = nil
	// Query the database for post data
	var rows *sql.Rows
	if !Filtered {
		rows, err = Postsdb.Query("SELECT post_id, user_id, username,userImg, title, content , image, category, like, dislike, timestamp FROM posts")
		if err != nil {
			return nil, fmt.Errorf("failed to query database: %v", err)
		}
	} else {
		if Type == "ProfileFilter" {
			filterString := strings.Join(Filter, ",")
			rows, err = Postsdb.Query("SELECT post_id, user_id, username, userImg, title, content, image, category, like, dislike, timestamp FROM posts WHERE post_id IN (" + filterString + ")")
			if err != nil {
				return nil, fmt.Errorf("failed to query database: %v", err)
			}
		} else if Type == "Categories" {

			placeholders := make([]string, len(Filter))
			args := make([]interface{}, len(Filter))

			for i, category := range Filter {
				placeholders[i] = "category LIKE ?"
				args[i] = "%" + category + "%"
			}

			splitString := strings.Split(Filter[0], " ")
			f := splitString[0]
			if f == "News" {
				args[0] = "%News%"
			}

			placeholderString := strings.Join(placeholders, " OR ")

			query := fmt.Sprintf(`SELECT post_id, user_id, username, userImg, title, content, image, category, like, dislike, timestamp 
    FROM posts
    WHERE %s`, placeholderString)

			// Execute the query
			rows, err = Postsdb.Query(query, args...)
			if err != nil {
				log.Println("Error querying the database:", err)
			}

		}
	}
	defer rows.Close()
	// Iterate over the rows and populate the Posts slice
	for rows.Next() {
		var post Post
		var category string

		err := rows.Scan(&post.PostId, &post.UserId, &post.UserName, &post.UserImg, &post.Title, &post.Content, &post.Image, &category, &post.Like, &post.Dislike, &post.Time)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}

		// Split the category string into a slice of categories
		post.Category = strings.Split(category, " ")

		var commentArray []Comment
		var comment Comment
		crow, err := Commentsdb.Query("SELECT comment_id, userName, text, ProfileImage, timestamp, CLike, CDislike FROM comments WHERE post_id = ? ", post.PostId)
		if err != nil {
			return nil, fmt.Errorf("failed to query database: %v", err)
		}
		defer crow.Close()
		for crow.Next() {
			err := crow.Scan(&comment.CommentId, &comment.UserName, &comment.Text, &comment.ProfileImage, &comment.Time, &comment.CLike, &comment.CDislike)
			if err != nil {
				log.Println("Error scanning row:", err)
				continue
			}
			commentArray = append(commentArray, comment)
		}

		post.Comments = commentArray

		// Append the post to the Posts slice
		Posts = append(Posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return Posts, nil
}

func HandleLikeRequest(w http.ResponseWriter, r *http.Request, user Account, request Request) {
	// Get the current like value
	var currentLike int
	err = Postsdb.QueryRow("SELECT "+request.Type+" FROM posts WHERE post_id = ?", request.ID).Scan(&currentLike)
	if err != nil {
		log.Fatal(err)
	}
	if request.Checked == true {
		// Execute the update statement
		_, err := Postsdb.Exec("UPDATE posts SET "+request.Type+" = ? WHERE post_id = ?", currentLike+1, request.ID)
		if err != nil {
			log.Fatal(err)
		}

		_, err = Postsdb.Exec(`INSERT INTO LikedPosts (user_id, post_id, type) VALUES (?, ?, ?)`, user.Id, request.ID, request.Type)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Handle unchecked state
		_, err = Postsdb.Exec("UPDATE posts SET "+request.Type+" = ? WHERE post_id = ?", currentLike-1, request.ID)
		if err != nil {
			log.Fatal(err)
		}

		_, err = Postsdb.Exec(`DELETE FROM LikedPosts WHERE user_id = ? AND post_id = ? AND type = ?`, user.Id, request.ID, request.Type)
	}
}

func GetLikedPosts(id int) ([]string, []string) {
	var LikedPosts []string
	var LikedPostsId []string

	// Query the database for post data
	rows, err := Postsdb.Query("SELECT post_id, type FROM LikedPosts WHERE user_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Iterate over the rows and populate the Posts slice
	for rows.Next() {
		var PostId int
		var Type string

		err := rows.Scan(&PostId, &Type)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		if Type == "Like" {
			LikedPosts = append(LikedPosts, "Like_"+strconv.Itoa(PostId))
			LikedPostsId = append(LikedPostsId, strconv.Itoa(PostId))

		} else {
			LikedPosts = append(LikedPosts, "Dislike_"+strconv.Itoa(PostId))
		}
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	return LikedPostsId, LikedPosts
}

// To reverse the posts array
func reverseArray(arr []Post) {
	length := len(arr)
	for i := 0; i < length/2; i++ {
		arr[i], arr[length-i-1] = arr[length-i-1], arr[i]
	}
}

func HandleDeleteRequest(request Request) {
	PostId, err := strconv.Atoi(request.ID)
	if err != nil {
		log.Fatal(err)
	}
	_, err = Postsdb.Exec("DELETE FROM posts WHERE post_id = ?", PostId)
	if err != nil {
		log.Fatal(err)
	}
	_, err = Postsdb.Exec("DELETE FROM LikedPosts WHERE post_id = ?", PostId)
	if err != nil {
		log.Fatal(err)
	}

}

func GetCreatedPosts(UserId int) []string {
	var CreatedPosts []string

	// Query the database for post data
	rows, err := Postsdb.Query("SELECT post_id FROM posts WHERE user_id = ?", UserId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Iterate over the rows and populate the Posts slice
	for rows.Next() {
		var PostId int

		err := rows.Scan(&PostId)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		CreatedPosts = append(CreatedPosts, strconv.Itoa(PostId))
	}
	if err = rows.Err(); err != nil {
		//Handle Error
	}
	return CreatedPosts
}
