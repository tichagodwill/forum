package forum
import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"forum/Error"
)
type Payload struct {
	Message string `json:"message"`
	PostId  int `json:"postId"`
	Username string `json:"username"`
	ProfileImage string `json:"profileImage"`
}
type CommentResponse struct {
	CommentId int64 `json:"CommentId"`
}
func CommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error.RenderErrorPage(w , http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var payload Payload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		Error.RenderErrorPage(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	message := payload.Message
	postId := payload.PostId
	username := payload.Username
	profileImage := payload.ProfileImage
	//userId := payload.UserId
	// Create a new comment
	comment := Comment{
		//PostId: PostId,
		UserName: username,
		Text: message,
		ProfileImage: profileImage,
	}
	// Insert the comment into the database
	result, err := Commentsdb.Exec("INSERT INTO comments (user_id, userName, post_id, text, ProfileImage) VALUES (?,?, ?, ?,?)",
		user.Id, comment.UserName, postId, comment.Text, profileImage)
	if err != nil {
		log.Fatal(err)
	}
	CommentId, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	
	// Create a CommentResponse object
	response := CommentResponse{
		CommentId: CommentId,
	}
	
	// Set the response content type and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(response)
	if err != nil {
		Error.RenderErrorPage(w, http.StatusInternalServerError, "Error encoding response")
		return
	}
}
func CommentLikeHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error.RenderErrorPage(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var request Request
	//var user Account
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		Error.RenderErrorPage(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	var user Account
	// Query the user data based on the ID
	user = GetUserData(w, r)
	var currentLike int
	err = Commentsdb.QueryRow("SELECT C"+request.Type+" FROM comments WHERE comment_id = ?", request.ID).Scan(&currentLike)
	if err != nil {
		log.Fatal(err)
	}
	if request.Checked == true {
		// Execute the update statement
		_, err := Commentsdb.Exec("UPDATE comments SET C"+request.Type+" = ? WHERE comment_id = ?", currentLike+1, request.ID)
		if err != nil {
			log.Fatal(err)
		}
		_, err = Commentsdb.Exec(`INSERT INTO Likedcomments (user_id, comment_id, type) VALUES (?, ?, ?)`, user.Id, request.ID, request.Type)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Handle unchecked state
		_, err = Commentsdb.Exec("UPDATE comments SET C"+request.Type+" = ? WHERE comment_id = ?", currentLike-1, request.ID)
		if err != nil {
			log.Fatal(err)
		}
		_, err = Commentsdb.Exec(`DELETE FROM Likedcomments WHERE user_id = ? AND comment_id = ? AND type = ?`, user.Id, request.ID, request.Type)
}
}
func GetLikedComments(id int) []string {
	var LikedComments []string
	// Query the database for post data
	rows, err := Commentsdb.Query("SELECT comment_id, type FROM Likedcomments WHERE user_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	// Iterate over the rows and populate the Posts slice
	for rows.Next() {
		var CommentId int
		var Type string
		err := rows.Scan(&CommentId, &Type)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		if Type == "Like" {
			LikedComments = append(LikedComments, "CLike_"+strconv.Itoa(CommentId))
		} else {
			LikedComments = append(LikedComments, "CDislike_"+strconv.Itoa(CommentId))
		}
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return LikedComments
}