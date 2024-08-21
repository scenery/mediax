package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/scenery/mediax/handlers"

	"github.com/google/uuid"
)

func handleSubject(w http.ResponseWriter, r *http.Request) {
	var err error
	pathParts := strings.Split(r.URL.Path[1:], "/")
	subjectType := pathParts[0]
	uuidStr := pathParts[1]
	if uuidStr == "" {
		http.Redirect(w, r, "/"+subjectType, http.StatusSeeOther)
		return
	} else {
		_, err = uuid.Parse(uuidStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	if len(pathParts) == 2 {
		handleViewSubject(w, uuidStr)
		return
	} else if len(pathParts) == 3 {
		switch pathParts[2] {
		case "edit":
			handleEditSubject(w, r, uuidStr)
			return
		case "delete":
			handleDeleteSubject(w, r, subjectType, uuidStr)
			return
		default:
			handleError(w, "404 page not found", "home", 404)
			return
		}
	} else {
		handleError(w, "404 page not found", "home", 404)
		return
	}
}

func handleViewSubject(w http.ResponseWriter, uuidStr string) {
	subject, err := handlers.GetSubject(uuidStr)
	if err != nil {
		handleError(w, fmt.Sprint(err), "home", 404)
		return
	}

	data := processSingleHTML(subject.Title, 1, subject)
	renderPage(w, "single.html", data)
}

func handleEditSubject(w http.ResponseWriter, r *http.Request, uuidStr string) {
	switch r.Method {
	case http.MethodGet:
		subject, err := handlers.GetSubject(uuidStr)
		if err != nil {
			handleError(w, fmt.Sprint(err), "home", 500)
			return
		}
		data := processSingleHTML("编辑 "+subject.Title, 2, subject)
		renderPage(w, "manage.html", data)
		return
	case http.MethodPost:
		subjectURL, err := handlers.ManageSubject(w, r, uuidStr)
		if err != nil {
			handleError(w, fmt.Sprint(err), "home", 500)
			return
		}
		http.Redirect(w, r, subjectURL, http.StatusSeeOther)
		return
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func handleDeleteSubject(w http.ResponseWriter, r *http.Request, subjectType, uuidStr string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		handleError(w, fmt.Sprint(err), "home", 500)
		return
	}
	deleteConfirm := r.FormValue("confirm_delete")
	if deleteConfirm == "purge-it" {
		err = handlers.ManageDelSubject(uuidStr, subjectType)
		if err != nil {
			handleError(w, fmt.Sprint(err), "home", 500)
			return
		}
		http.Redirect(w, r, "/"+subjectType, http.StatusSeeOther)
		return
	}
}
