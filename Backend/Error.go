package forum

import (
    "net/http"
    // "log"
    // "os"
)

// Function to serve error files
func ServeErrorFiles() {
    // Print current working directory
    // cwd, err := os.Getwd()
    // if err != nil {
    //     log.Println("Error getting current working directory:", err)
    // } else {
    //     log.Println("Current working directory:", cwd)
    // }

    // Serve error files from the "Error" directory
    fs := http.FileServer(http.Dir("Error"))

    // Route requests to error files under the "/error/" URL prefix
    http.Handle("/error/", http.StripPrefix("/error/", fs))
}

// Function to handle errors by rendering appropriate error pages
func ErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int, errorMsg string) {
    switch statusCode {
    case http.StatusBadRequest:
        http.ServeFile(w, r, "Error/error400.html")
    case http.StatusNotFound:
        http.ServeFile(w, r, "Error/error404.html") 
    case http.StatusInternalServerError:
        http.ServeFile(w, r, "Error/error500.html") 
    default:
        w.Write([]byte(errorMsg))
    }
}

// Function to handle bad request errors
func HandleBadRequest(w http.ResponseWriter, r *http.Request) {
    ErrorHandler(w, r, http.StatusBadRequest, "Bad Request")
}

// Function to handle not found errors
func HandleNotFound(w http.ResponseWriter, r *http.Request) {
    // log.Println("Handling not found error") 
    ErrorHandler(w, r, http.StatusNotFound, "Page Not Found")
}

// Function to handle internal server errors
func HandleInternalServerError(w http.ResponseWriter, r *http.Request) {
    ErrorHandler(w, r, http.StatusInternalServerError, "Internal Server Error")
}

func ServeStaticFiles() {
    fs := http.FileServer(http.Dir("Error"))
    http.Handle("/Error/", http.StripPrefix("/Error/", fs))
}

func init() {
    ServeErrorFiles()
    ServeStaticFiles()
}
