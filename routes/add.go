package routes

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/scenery/mediax/dataops"
	"github.com/scenery/mediax/handlers"
	"github.com/scenery/mediax/helpers"
	"github.com/scenery/mediax/models"
)

func handleAdd(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		queryParams := r.URL.Query()
		subjectType := queryParams.Get("subject_type")
		data := map[string]interface{}{
			"PageTitle":   "添加",
			"SubjectType": subjectType,
		}
		renderPage(w, "add.html", data)
	case http.MethodPost:
		err := handleAddMethod(w, r)
		if err != nil {
			handleError(w, fmt.Sprint(err), "add", 500)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleAddMethod(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	subjectType := r.FormValue("subject_type")
	externalURL := strings.TrimSpace(r.FormValue("external_url"))

	if externalURL == "" {
		manualAdd(w, subjectType)
		return nil
	}

	err = handlers.CheckSubjectExistence(externalURL)
	if err != nil {
		return err
	}

	return autoAdd(w, subjectType, externalURL)
}

func manualAdd(w http.ResponseWriter, subjectType string) {
	var subject models.Subject
	subject.SubjectType = subjectType
	subject.Status = 1
	subject.Rating = 0
	subject.MarkDate = time.Now().Format("2006-01-02")
	data := processSingleHTML("添加"+helpers.GetSubjectType(subjectType), 3, subject)
	renderPage(w, "manage.html", data)
}

func autoAdd(w http.ResponseWriter, subjectType, externalURL string) error {
	subjectID, apiTarget, err := processExternalURL(externalURL, subjectType)
	if err != nil {
		return err
	}

	subject, err := dataops.FetchMediaInfo(subjectType, subjectID, apiTarget, externalURL)
	if err != nil {
		return err
	}

	subject.ExternalURL = externalURL
	subject.SubjectType = subjectType
	subject.Status = 1
	subject.Rating = 0
	subject.MarkDate = time.Now().Format("2006-01-02")
	data := processSingleHTML("添加"+helpers.GetSubjectType(subjectType), 4, subject)
	renderPage(w, "manage.html", data)
	return nil
}

func handleAddSubject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	subjectURL, err := handlers.ManageSubject(w, r, "")
	if err != nil {
		handleError(w, fmt.Sprint(err), "add", 500)
		return
	}

	http.Redirect(w, r, subjectURL, http.StatusSeeOther)
}

func processExternalURL(externalURL string, subjectType string) (string, string, error) {
	var subjectID string
	var apiTarget string

	pattern := regexp.MustCompile(`^https://(?:(?:www|book|movie)\.douban\.com|(?:bgm|bangumi)\.tv)/(?:game|subject)/(\d+)/?$`)
	matched := pattern.MatchString(externalURL)
	if !matched {
		return "", "", errors.New("bad request: invalid link format")
	}

	matches := pattern.FindStringSubmatch(externalURL)
	if len(matches) <= 1 {
		return "", "", errors.New("url does not contain a valid subject ID")
	}
	subjectID = matches[1]

	u, err := url.Parse(externalURL)
	if err != nil {
		return "", "", err
	}
	urlHost := strings.Split(u.Hostname(), ":")[0]

	var validHosts []string
	switch subjectType {
	case "book":
		validHosts = []string{"book.douban.com", "bgm.tv", "bangumi.tv"}
	case "movie":
		validHosts = []string{"movie.douban.com", "bgm.tv", "bangumi.tv"}
	case "tv":
		validHosts = []string{"movie.douban.com", "bgm.tv", "bangumi.tv"}
	case "anime":
		validHosts = []string{"movie.douban.com", "bgm.tv", "bangumi.tv"}
	case "game":
		validHosts = []string{"www.douban.com", "bgm.tv", "bangumi.tv"}
	default:
		return "", "", errors.New("unknown subject type")
	}

	isValidHost := false
	for _, host := range validHosts {
		if urlHost == host {
			isValidHost = true
			break
		}
	}
	if !isValidHost {
		return "", "", errors.New("invalid URL for the given subject type. Supported hosts: " + strings.Join(validHosts, ", "))
	}

	if strings.Contains(externalURL, "douban") {
		apiTarget = "douban"
	} else {
		apiTarget = "bangumi"
	}

	return subjectID, apiTarget, nil
}
