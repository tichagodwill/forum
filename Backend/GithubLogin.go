package forum

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var githubOAuthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/github_callback",
	ClientID:     "",
	ClientSecret: "",
	Scopes:       []string{"user:email"},
	Endpoint:     github.Endpoint,
}

var oauthStateString = "random" // A random string for security purposes.

func GithubLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect to the Github login page
	http.Redirect(w, r, githubOAuthConfig.AuthCodeURL(oauthStateString), http.StatusFound)
}

func GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != oauthStateString {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}
	// Get the authorization code from the query string
	code := r.FormValue("code")
	fmt.Println("CODE: ", code)
	// Exchange the authorization code for an access token
	token, err := githubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Token: ", token)
	// Get the user's profile information
	// client := githubOAuthConfig.Client(context.Background(), token)
	// response, err := client.Get("https://api.github.com/user")
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// response.Body.Close()

	// userData, err := ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// output := getGithubData(token.AccessToken)
	// fmt.Println("OUTPUT: ")
	// fmt.Println(output)
	// // fmt.Println(string(userData))
	// fmt.Fprint(w, output)

	// Get request to a set URL
	req, reqerr := http.NewRequest(
		"GET",
		"https://api.github.com/user",
		nil,
	)
	if reqerr != nil {
		log.Panic("API Request creation failed")
	}

	// Set the Authorization header before sending the request
	// Authorization: token XXXXXXXXXXXXXXXXXXXXXXXXXXX
	authorizationHeaderValue := fmt.Sprintf("token %s", token.AccessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	// Make the request
	response, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		log.Panic("Request failed")
	}

	var userInfo struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		// VerifiedEmail bool   `json:"verified_email"`
		// Name string `json:"name"`
		// GivenName     string `json:"given_name"`
		// Picture       string `json:"picture"`
	}

	//parse the response body
	if err := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		log.Printf("Could not parse response: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// fmt.Println("OUTPUT: ")
	// fmt.Println(userInfo)
	// fmt.Fprintf(w, "User Info: %+v\n", userInfo)

	var tmpl = template.Must(template.ParseFiles("./Pages/Login.html"))
	exists, err := AccountGithubExists(userInfo.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if exists {
		fmt.Println("Signing in with github")
		ID, err := GetGithubAccountID("", userInfo.Login, userInfo.ID)
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
		fmt.Println("Signing up with github")
		ID, errorMessage, err := AddGithubAccount("", userInfo.Login, userInfo.ID)
		if errorMessage != "" {
			Template(w, tmpl, errorMessage)
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

func AccountGithubExists(username string) (bool, error) {
	query := "SELECT COUNT(*) FROM accounts WHERE Username = ?"
	row := Accountsdb.QueryRow(query, username)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func getGithubData(accessToken string) string {
	// Get request to a set URL
	req, reqerr := http.NewRequest(
		"GET",
		"https://api.github.com/user",
		nil,
	)
	if reqerr != nil {
		log.Panic("API Request creation failed")
	}

	// Set the Authorization header before sending the request
	// Authorization: token XXXXXXXXXXXXXXXXXXXXXXXXXXX
	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	// Make the request
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		log.Panic("Request failed")
	}

	// Read the response as a byte slice
	respbody, _ := ioutil.ReadAll(resp.Body)

	// Convert byte slice to string and return
	return string(respbody)
}

func AddGithubAccount(email, username string, githubUserId int) (int64, string, error) {
	// Convert ID int to string
	IDstring := strconv.Itoa(githubUserId)
	errorMessage := ""
	insertQuery := "INSERT INTO accounts (Email, GoogleUserID, Username) VALUES (?, ?, ?) ON CONFLICT(GoogleUserID) DO UPDATE SET Email=excluded.Email, Username=excluded.Username"

	result, err := Accountsdb.Exec(insertQuery, email, IDstring, username)
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

func GetGithubAccountID(email, username string, githubUserId int) (int64, error) {
	// Convert ID int to string
	IDstring := strconv.Itoa(githubUserId)
	// Prepare the SQL statement
	query := "SELECT id FROM accounts WHERE Email = ? AND Username = ? AND GoogleUserID = ?"
	stmt, err := Accountsdb.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Execute the query and get the account ID
	var accountID int64
	err = stmt.QueryRow(email, username, IDstring).Scan(&accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Account does not exist
			return 0, nil
		}
		return 0, err
	}

	return accountID, nil
}
