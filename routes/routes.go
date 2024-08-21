package routes

import (
	"log"
	"net/http"
	"os"

	"github.com/scenery/mediax/config"
	"github.com/scenery/mediax/web"
)

func setupRoutes() {
	var err error
	staticFS, err = web.GetStaticFileSystem()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	if _, err := os.Stat(config.ImageDir); os.IsNotExist(err) {
		err := os.MkdirAll(config.ImageDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create image directory <%s>: %v", config.ImageDir, err)
		}
		log.Printf("Image directory <%s> did not exist, it has been created automatically", config.ImageDir)
	}
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(config.ImageDir))))

	http.HandleFunc("/", redirectToHome)
	http.HandleFunc("/home", handleHomePage)

	http.HandleFunc("/book", handleCategory)
	http.HandleFunc("/movie", handleCategory)
	http.HandleFunc("/tv", handleCategory)
	http.HandleFunc("/anime", handleCategory)
	http.HandleFunc("/game", handleCategory)

	http.HandleFunc("/book/", handleSubject)
	http.HandleFunc("/movie/", handleSubject)
	http.HandleFunc("/tv/", handleSubject)
	http.HandleFunc("/anime/", handleSubject)
	http.HandleFunc("/game/", handleSubject)

	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/add/subject", handleAddSubject)

	http.HandleFunc("/search", handleSearch)
}
