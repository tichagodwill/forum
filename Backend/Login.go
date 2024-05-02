package forum

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"time"
)

func Login(w http.ResponseWriter, r *http.Request) {
	var tmpl = template.Must(template.ParseFiles("./Pages/Login.html"))

	if r.Method == "GET" {
		// Display the account creation form
		err := tmpl.Execute(w, map[string]interface{}{
			"Error": "",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		var account Account
		account.Email = r.FormValue("email")
		account.Username = r.FormValue("username")
		account.Password = r.FormValue("password")
		action := r.FormValue("action")
		

		switch action {
		case "login":

			// check if account Exists
			exists, err := AccountExists(account.Email)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if account.Email == "" || account.Username == "" || account.Password == "" {
				errorMessage := "Please Fill All The Boxes"
				Template(w, tmpl, errorMessage)
			} else if !exists {
				errorMessage := "This Account Does Not Exist. Please Try Again"
				Template(w, tmpl, errorMessage)
			} else {
				//HandlerLoginCookies(w, r, account)
				var password string
				ID, err := GetAccountID(account.Email, account.Username)
				if ID == 0 {
					errorMessage := "This Account Does Not Exist. Please Try Again"
					Template(w, tmpl, errorMessage)
				} else {
					er := Accountsdb.QueryRow("SELECT Password FROM accounts WHERE id = ?", ID).Scan(&password)
					if !decryption(account.Password, password) {
						errorMessage := "worng password"
						Template(w, tmpl, errorMessage)
						if er != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
					} else {
						sessionID, expt := Cookies(w)
						SessionID(ID, sessionID, expt)
						guest = false
						http.Redirect(w, r, "/HomePage", http.StatusFound)

					}
				}
			}
		case "signup":
			http.Redirect(w, r, "/SignUp", http.StatusFound)
		case "guest":
			guest = true
			http.Redirect(w, r, "/HomePage", http.StatusFound)
		}

	}
}

func encrytion(password string) string {
	// Generate a bcrypt hash of the password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the hash to a string and store it in the database
	hashedPassword := string(hash)
//	fmt.Println("Hashed password:", hashedPassword)
	return hashedPassword

}

func decryption(providedPassword string, hashedPassword string) bool{
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
	if err == nil {
		//fmt.Println("Password is correct")
		return true
	} else if err == bcrypt.ErrMismatchedHashAndPassword {
		//fmt.Println("Password is incorrect")
		return false
	} else {
		log.Fatal(err)
		return false
	}
}

func AccountExists(email string) (bool, error) {
	query := "SELECT COUNT(*) FROM accounts WHERE Email = ?"
	row := Accountsdb.QueryRow(query, email)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func SessionID(ID int64, sessionID string, expt time.Time) error {
	updateQuery := "UPDATE accounts SET SessionID = ?, Expiration = ? WHERE id = ?"
	_, err := Accountsdb.Exec(updateQuery, sessionID, expt, ID)
	if err != nil {
		return err
	}
	return nil
}

func GetAccountID(email, username string) (int64, error) {
	// Prepare the SQL statement
	query := "SELECT id FROM accounts WHERE Email = ? AND Username = ?"
	stmt, err := Accountsdb.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Execute the query and get the account ID
	var accountID int64
	err = stmt.QueryRow(email, username).Scan(&accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Account does not exist
			return 0, nil
		}
		return 0, err
	}

	return accountID, nil
}

func Cookies(w http.ResponseWriter) (string, time.Time) {
	expiration := time.Now().Add(24 * time.Hour) // Set the expiration time to 24 hours from now
	// Generate a unique session ID
	sessionID := generateSessionID()
	// Store the session ID in a cookie
	cookie := &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		MaxAge:   60 * 15,
		Expires:  expiration,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
	return sessionID, expiration
}
