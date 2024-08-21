package dataops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scenery/mediax/models"

	"gorm.io/gorm"
)

func joinStrings(elements []string) string {
	return strings.Join(elements, ", ")
}

func parseDate(dateStr string) string {
	dateFormats := []string{
		time.RFC3339, //  "2024-08-09T00:21:52+08:00"
		"2006-01-02 15:04:05",
	}
	for _, format := range dateFormats {
		parsedTime, err := time.Parse(format, dateStr)
		if err == nil {
			return parsedTime.Format("2006-01-02")
		}
	}
	return "Invalid date"
}

func getDoubanStatus(status string) int {
	switch status {
	case "mark":
		return 1
	case "doing":
		return 2
	case "done":
		return 3
	default:
		return 0
	}
}

func getCreator(directors []models.DoubanDirector) string {
	var names []string
	for _, director := range directors {
		names = append(names, director.Name)
	}
	return strings.Join(names, " / ")
}

func getFirstPubdate(elements []string) string {
	element := "unknown"
	if len(elements) > 0 {
		element = elements[0]
	}
	return element
}

func getBangumiType(itemType int) string {
	switch itemType {
	case 1:
		return "book"
	case 2:
		return "anime"
	case 4:
		return "game"
	case 6:
		return "tv"
	default:
		return "unknown"
	}
}

func getBangumiTitle(name, nameCN string) (string, string) {
	title := name
	var altTitle string

	if nameCN == "" || nameCN == name {
		altTitle = ""
	} else {
		title = nameCN
		altTitle = name
	}
	return title, altTitle
}

func getBangumiStatus(status int) int {
	switch status {
	case 2:
		return 3
	case 3:
		return 2
	default:
		return status
	}
}

func getExistingRecords(db *gorm.DB) map[string]models.Subject {
	existingRecords := make(map[string]models.Subject)
	var existingSubjects []models.Subject

	if err := db.Model(&models.Subject{}).Select("external_url, uuid").Find(&existingSubjects).Error; err != nil {
		fmt.Printf("failed to query existing records: %v\nThis is an expected error, continuing processing...\n", err)
		return existingRecords
	}

	for _, subject := range existingSubjects {
		existingRecords[subject.ExternalURL] = subject
	}

	return existingRecords
}

func getExistingImages(baseDir string) map[string]bool {
	imageCache := make(map[string]bool)

	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\nThis is an expected error, continuing processing...\n", path, err)
			return nil
		}
		if !info.IsDir() {
			relativePath, err := filepath.Rel(baseDir, path)
			if err != nil {
				fmt.Printf("Error getting relative path for %s: %v\nThis is an expected error, continuing processing...\n", path, err)
				return nil
			}
			imageCache[relativePath] = true
		}
		return nil
	})

	return imageCache
}
