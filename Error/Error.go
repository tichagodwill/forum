package Error
import (
	"html/template"
	"net/http"
)
type CustomError struct {
	Code    int
	Message string
}
func RenderErrorPage(w http.ResponseWriter, Code int, Message string) {
	w.WriteHeader(Code)
	tmpl, _ := template.ParseFiles("Error/Error.html")
	er := CustomError{
		Code:    Code,
		Message: Message,
	}
	_ = tmpl.Execute(w, er)
}