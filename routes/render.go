package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/scenery/mediax/config"
	"github.com/scenery/mediax/dataops"
	"github.com/scenery/mediax/models"
)

func renderPage(w http.ResponseWriter, contentTemplate string, data interface{}) {
	pageTemplates, err := baseTemplates.Clone()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to clone base templates: %v", err), http.StatusInternalServerError)
		return
	}

	pageTemplates, err = pageTemplates.ParseFS(tmplFS, contentTemplate)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse templates: %v", err), http.StatusInternalServerError)
		return
	}

	err = pageTemplates.ExecuteTemplate(w, "baseof.html", data)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to render page: %v", err), http.StatusInternalServerError)
		return
	}
}

func handleError(w http.ResponseWriter, errorMessage, targetURL string, statusCode int) {
	data := struct {
		ErrorMessage string
		TargetPath   string
	}{
		ErrorMessage: errorMessage,
		TargetPath:   targetURL,
	}

	errorHTML, err := template.ParseFS(tmplFS, "error.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)
	err = errorHTML.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}
}

func processSingleHTML(pageTitle string, manageType int, subject models.Subject) models.SubjectView {
	imageURL := getImageURL(manageType, subject.HasImage, subject.SubjectType, subject.UUID, subject.ExternalURL)

	labels := processSubjectLabel(subject.SubjectType, subject.Status)
	statusText := labels["statusText"]
	creatorLabel := labels["creatorLabel"]
	pressLabel := labels["pressLabel"]
	pubDateLabel := labels["pubDateLabel"]
	summaryLabel := labels["summaryLabel"]

	processedSubject := models.SubjectView{
		PageTitle:    pageTitle,
		ManageType:   manageType,
		CreatorLabel: creatorLabel,
		PressLabel:   pressLabel,
		PubDateLabel: pubDateLabel,
		StatusText:   statusText,
		SummaryLabel: summaryLabel,
		ImageURL:     imageURL,
		Subject:      subject,
	}

	if subject.Rating != 0 {
		processedSubject.RatingStar = subject.Rating * 5
	}

	if subject.ExternalURL != "" {
		processedSubject.ExternalURLIcon = getExternalURLIcon(subject.ExternalURL)
	}

	return processedSubject
}

func processCategoryHTML(subjects []models.SubjectSummary) []models.CategoryViewItem {
	var processedSubjects []models.CategoryViewItem

	for _, subject := range subjects {
		imageURL := getImageURL(0, subject.HasImage, subject.SubjectType, subject.UUID, "")

		labels := processSubjectLabel(subject.SubjectType, subject.Status)
		statusText := labels["statusText"]
		creatorLabel := labels["creatorLabel"]
		pressLabel := labels["pressLabel"]
		pubDateLabel := labels["pubDateLabel"]

		processedSubjects = append(processedSubjects, models.CategoryViewItem{
			SubjectType:  subject.SubjectType,
			SubjectURL:   fmt.Sprintf("/%s/%s", subject.SubjectType, subject.UUID),
			Title:        subject.Title,
			AltTitle:     subject.AltTitle,
			Creator:      subject.Creator,
			Press:        subject.Press,
			PubDate:      subject.PubDate,
			MarkDate:     subject.MarkDate,
			Rating:       subject.Rating,
			StatusText:   statusText,
			CreatorLabel: creatorLabel,
			PressLabel:   pressLabel,
			PubDateLabel: pubDateLabel,
			ImageURL:     imageURL,
		})
	}

	return processedSubjects
}

func processHomeHTML(subjects []models.SubjectSummary) []models.HomeViewItem {
	var processedSubjects []models.HomeViewItem

	for _, subject := range subjects {
		imageURL := getImageURL(0, subject.HasImage, subject.SubjectType, subject.UUID, "")

		var isDoing bool
		if subject.Status == 2 {
			isDoing = true
		} else {
			isDoing = false
		}

		processedSubjects = append(processedSubjects, models.HomeViewItem{
			SubjectURL: fmt.Sprintf("/%s/%s", subject.SubjectType, subject.UUID),
			Title:      subject.Title,
			MarkDate:   subject.MarkDate,
			IsDoing:    isDoing,
			ImageURL:   imageURL,
		})
	}

	return processedSubjects
}

func processSubjectLabel(subjectType string, status int) map[string]string {
	result := make(map[string]string)

	statusType := "看"
	creatorLabel := "导演"
	pressLabel := "制片国家/地区"
	pubDateLabel := "上映日期"
	summaryLabel := "剧情简介"

	switch subjectType {
	case "book":
		statusType = "读"
		creatorLabel = "作者"
		pressLabel = "出版社"
		pubDateLabel = "出版日期"
		summaryLabel = "内容简介"
	case "anime":
		pressLabel = "动画制作"
		pubDateLabel = "放送日期"
	case "game":
		statusType = "玩"
		creatorLabel = "开发团队"
		pressLabel = "发行公司"
		pubDateLabel = "发行日期"
		summaryLabel = "游戏简介"
	}

	var statusText string
	switch status {
	case 1:
		statusText = fmt.Sprintf("想%s", statusType)
	case 2:
		statusText = fmt.Sprintf("在%s", statusType)
	case 3:
		statusText = fmt.Sprintf("%s过", statusType)
	case 4:
		statusText = "搁置"
	case 5:
		statusText = "抛弃"
	default:
		statusText = "未知"
	}

	result["statusText"] = statusText
	result["creatorLabel"] = creatorLabel
	result["pressLabel"] = pressLabel
	result["pubDateLabel"] = pubDateLabel
	result["summaryLabel"] = summaryLabel

	return result
}

func getImageURL(manageType, hasImage int, subjectType, uuid, externalURL string) string {
	imgURL := "/static/default-cover.jpg"
	if hasImage == 1 {
		imgURL = fmt.Sprintf("/%s/%s/%s.jpg", config.ImageDir, subjectType, uuid)
	}
	if manageType == 4 {
		imageName, err := dataops.PreDownloadImageName(externalURL)
		if err == nil {
			imgURL = fmt.Sprintf("/%s/temp/%s", config.ImageDir, imageName)
		}
	}
	return imgURL
}

func getExternalURLIcon(externalURL string) template.HTML {
	switch {
	case strings.Contains(externalURL, "douban.com"):
		return template.HTML(fmt.Sprintf(`<a class="subject-outlink link-douban" href="%s" target="_blank" rel="noopener noreferrer">豆瓣</a>`, externalURL))
	case strings.Contains(externalURL, "bgm.tv"):
		return template.HTML(fmt.Sprintf(`<a class="subject-outlink link-bangumi" href="%s" target="_blank" rel="noopener noreferrer">Bangumi</a>`, externalURL))
	default:
		return template.HTML(fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer">%s</a>`, externalURL, externalURL))
	}
}
