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

func handleCategory(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimPrefix(r.URL.Path, "/")

	pageSize := config.PageSize
	page, err := helpers.StringToInt(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	status, err := helpers.StringToInt(r.URL.Query().Get("status"))
	if err != nil || status < 0 || status > 5 {
		status = 0
	}

	statusCounts, err := handlers.GetStatusCounts(category)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to calculate totals for %s: %v", category, err)
		handleError(w, errorMessage, "home", 500)
		return
	}

	var totalPages int
	getTotalPages := func(count int64) int {
		return int(math.Ceil(float64(count) / float64(pageSize)))
	}
	switch status {
	case 0:
		totalPages = getTotalPages(statusCounts.All)
	case 1:
		totalPages = getTotalPages(statusCounts.Todo)
	case 2:
		totalPages = getTotalPages(statusCounts.Doing)
	case 3:
		totalPages = getTotalPages(statusCounts.Done)
	case 4:
		totalPages = getTotalPages(statusCounts.OnHold)
	case 5:
		totalPages = getTotalPages(statusCounts.Dropped)
	}

	subjects, err := handlers.GetSubjectsByType(category, status, page, pageSize)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to get %s list: %v", category, err)
		handleError(w, errorMessage, "home", 500)
		return
	}

	data := models.CategoryView{
		PageTitle:    helpers.GetSubjectType(category),
		Category:     category,
		Status:       status,
		StatusCounts: statusCounts,
		Page:         page,
		Subjects:     processCategoryHTML(subjects),
		TotalPages:   totalPages,
	}

	renderPage(w, "category.html", data)
}
