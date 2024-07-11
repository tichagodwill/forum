package forum

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Google Login

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/google_callback",
	ClientID:     "",
	ClientSecret: "",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile", "openid"},
	Endpoint:     google.Endpoint,
}

// var oauthStateString = "random-string" // Use a more secure random generator in production
// Generate a random state string
func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(24 * time.Hour)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)
	return state
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	oauthStateString := generateStateOauthCookie(w)
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	var tmpl = template.Must(template.ParseFiles("./Pages/Login.html"))
	cookie, err := r.Cookie("oauthstate")
	if err != nil {
		errorMessage := "Google Login Failed. Please Try Again"
		Template(w, tmpl, errorMessage)
		return
		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// return
	}

	if r.FormValue("state") != cookie.Value {
		errorMessage := "Google Login Failed. Please Try Again"
		Template(w, tmpl, errorMessage)
		return
		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Printf("Could not get token: %s\n", err.Error())
		errorMessage := "Google Login Failed. Please Try Again"
		Template(w, tmpl, errorMessage)
		return
		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// return
	}
	// fmt.Println("Token: ", token.AccessToken)
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Could not create request: %s\n", err.Error())
		errorMessage := "Google Login Failed. Please Try Again"
		Template(w, tmpl, errorMessage)
		return
		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// return
	}
	defer response.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		// VerifiedEmail bool   `json:"verified_email"`
		// Name string `json:"name"`
		// GivenName     string `json:"given_name"`
		// Picture       string `json:"picture"`
	}

	//parse the response body
	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		log.Printf("Could not parse response: %s\n", err.Error())
		errorMessage := "Google Login Failed. Please Try Again"
		Template(w, tmpl, errorMessage)
		return
		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// return
	}

	// fmt.Fprintf(w, "User Info: %+v\n", userInfo)
	// print to console
	fmt.Println(userInfo)

	// Account creation and login
	// var tmpl = template.Must(template.ParseFiles("./Pages/Login.html"))
	exists, err := AccountExists(userInfo.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// if exist redirect to HomePage

	// var account Account
	// account.Email = userInfo.Email
	// split at @
	userInfoUsername := strings.Split(userInfo.Email, "@")[0]
	//account.Password = "" //userInfo.ID

	if exists {
		fmt.Println("Signing in with google")
		ID, err := GetGoogleAccountID(userInfo.Email, userInfoUsername, userInfo.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if ID == 0 {
			errorMessage := "This Account Does Not Exist. Please Try Again"
			Template(w, tmpl, errorMessage)
		} else {
			sessionID, expt := Cookies(w)
			SessionID(ID, sessionID, expt)
			guest = false
			http.Redirect(w, r, "/HomePage", http.StatusFound)
		}

	} else {
		fmt.Println("Signing up with google")
		ID, errorMessage, err := AddGoogleAccount(userInfo.Email, userInfo.ID, userInfoUsername)
		if errorMessage != "" {
			Template(w, tmpl, errorMessage)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sessionID, expt := Cookies(w)
		SessionID(ID, sessionID, expt)
		guest = false
		http.Redirect(w, r, "/HomePage", http.StatusFound)
	}
}

func AddGoogleAccount(email, googleUserId, username string) (int64, string, error) {
	errorMessage := ""
	insertQuery := "INSERT INTO accounts (Email, GoogleUserID, Username) VALUES (?, ?, ?) ON CONFLICT(GoogleUserID) DO UPDATE SET Email=excluded.Email, Username=excluded.Username"

	result, err := Accountsdb.Exec(insertQuery, email, googleUserId, username)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: accounts.Username") {
			errorMessage = "Username is already taken. Please choose a different username."
			fmt.Println(errorMessage)
			return 0, errorMessage, nil
		} else {
			errorMessage = "An error occurred while creating the account. Please try again later."
			return 0, errorMessage, nil
		}
	}

	Id, err := result.LastInsertId()
	if err != nil {
		return 0, errorMessage, err
	}
	return Id, errorMessage, nil
}

func GetGoogleAccountID(email, username, googleUserId string) (int64, error) {
	// Prepare the SQL statement
	query := "SELECT id FROM accounts WHERE Email = ? AND Username = ? AND GoogleUserID = ?"
	stmt, err := Accountsdb.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Execute the query and get the account ID
	var accountID int64
	err = stmt.QueryRow(email, username, googleUserId).Scan(&accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Account does not exist
			return 0, nil
		}
		return 0, err
	}

	return accountID, nil
}
