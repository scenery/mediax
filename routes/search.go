package routes

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/scenery/mediax/config"
	"github.com/scenery/mediax/handlers"
	"github.com/scenery/mediax/helpers"
	"github.com/scenery/mediax/models"
)

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) < 4 || len(query) > 32 {
		errorMessage := "search keyword must be between 4 and 32 characters (2 至 16 个汉字)"
		handleError(w, errorMessage, "home", 302)
		return
	}

	validType := map[string]bool{
		"all":   true,
		"book":  true,
		"movie": true,
		"tv":    true,
		"anime": true,
		"game":  true,
	}
	subjectType := r.URL.Query().Get("subject_type")
	if !validType[subjectType] {
		handleError(w, "Invalid subject type", "home", 400)
		return
	}

	pageSize := config.PageSize
	page, err := helpers.StringToInt(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	subjects, total, err := handlers.GetSearchResult(query, page, pageSize)
	if err != nil {
		handleError(w, fmt.Sprint(err), "add", 500)
		return
	}

	if subjectType != "all" {
		var filteredSubjects []models.SubjectSummary
		for _, subject := range subjects {
			if subject.SubjectType == subjectType {
				filteredSubjects = append(filteredSubjects, subject)
			}
		}
		subjects = filteredSubjects
		total = int64(len(filteredSubjects))
	}

	data := models.SearchView{
		PageTitle:   "搜索结果 " + query,
		Query:       query,
		QueryType:   subjectType,
		TotalCount:  total,
		CurrentPage: page,
		TotalPages:  int(math.Ceil(float64(total) / float64(pageSize))),
		Subjects:    processCategoryHTML(subjects),
	}

	renderPage(w, "search.html", data)
}
