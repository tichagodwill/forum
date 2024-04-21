package forum
import (
	"forum/Error"
	"log"
	"net/http"
	"time"
	_ "github.com/mattn/go-sqlite3"
)
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if guest {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// Get the session id from the cookie
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		Error.RenderErrorPage(w, http.StatusInternalServerError, "Error retrieving session cookie")
	}
	LogOutBySession(w, r, sessionCookie.Value)
}
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if guest {
			next.ServeHTTP(w, r)
		} else {
			// Get the session ID from the cookie
			sessionCookie, err := r.Cookie("session")
			if err != nil {
				// Redirect to the login page if the session ID cookie is missing
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			// Session cookie found, retrieve the session ID
			sessionID := sessionCookie.Value
			id, _ := getIDBySessionID(sessionID)
			if id == 0 {
				eror := LogOutBySession(w, r, sessionID)
				if eror != nil {
					Error.RenderErrorPage(w, http.StatusInternalServerError, "Error retrieving user ID")
					return
				}
			} else {
				var expirationTime time.Time
				// Check if the session ID exists in the database and is not expired
				err = Accountsdb.QueryRow("SELECT Expiration FROM accounts WHERE SessionID = ?", sessionID).
					Scan(&expirationTime)
				if err != nil {
					log.Fatal(err)
					return
				}
				if time.Now().After(expirationTime) {
					LogOutBySession(w, r, sessionID)
				}
				// Call the next handler if the session ID is valid
				next.ServeHTTP(w, r)
			}
		}
	}
}
