package routes

import (
	"fmt"
	"net/http"

	"github.com/scenery/mediax/handlers"
	"github.com/scenery/mediax/models"
)

func redirectToHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
}

func handleHomePage(w http.ResponseWriter, r *http.Request) {
	recentSubjects, err := handlers.GetRecentSubjects()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get recent subjects: %v", err), http.StatusInternalServerError)
		return
	}

	recentBooks := recentSubjects["book"]
	recentMovies := recentSubjects["movie"]
	recentTVs := recentSubjects["tv"]
	recentAnimes := recentSubjects["anime"]
	recentGames := recentSubjects["game"]

	data := models.HomeView{
		PageTitle:    "主页",
		FewBooks:     len(recentBooks) < 5,
		FewMovies:    len(recentMovies) < 5,
		FewTVs:       len(recentTVs) < 5,
		FewAnimes:    len(recentAnimes) < 5,
		FewGames:     len(recentGames) < 5,
		RecentBooks:  processHomeHTML(recentBooks),
		RecentMovies: processHomeHTML(recentMovies),
		RecentTVs:    processHomeHTML(recentTVs),
		RecentAnimes: processHomeHTML(recentAnimes),
		RecentGames:  processHomeHTML(recentGames),
	}

	renderPage(w, "index.html", data)
}
