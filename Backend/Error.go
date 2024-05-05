package forum

import (
	"html/template"
	"net/http"
)

// HandleBadRequest handles bad request errors
func HandleBadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	tmpl, err := template.ParseFiles("../Error/400-error.html")
	if err != nil {
		http.Error(w, "Bad request", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
