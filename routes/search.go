package routes

import (
	"fmt"
	"net/http"

	"github.com/scenery/mediax/handlers"
	"github.com/scenery/mediax/models"
)

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 4 || len(query) > 32 {
		errorMessage := "search keyword must be between 4 and 32 characters (2 至 16 个汉字)"
		targetURL := "home"
		handleError(w, errorMessage, targetURL, 302)
		return
	}

	subjects, total, err := handlers.GetSearchResult(query)
	if err != nil {
		handleError(w, fmt.Sprint(err), "add", 500)
		return
	}

	data := models.SearchView{
		PageTitle:  "搜索结果 " + query,
		TotalCount: total,
		Subjects:   processCategoryHTML(subjects),
	}

	renderPage(w, "search.html", data)
}
