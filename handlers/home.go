package handlers

import (
	"fmt"

	"github.com/scenery/mediax/cache"
	"github.com/scenery/mediax/database"
	"github.com/scenery/mediax/models"
)

func GetRecentSubjects() (map[string][]models.SubjectSummary, error) {
	db := database.GetDB()
	results := make(map[string][]models.SubjectSummary)
	subjectTypes := []string{"book", "movie", "tv", "anime", "game"}

	for _, subjectType := range subjectTypes {
		cacheKey := fmt.Sprintf("home:%s", subjectType)
		if cachedValue, found := cache.GetCache(cacheKey); found {
			results[subjectType] = cachedValue.([]models.SubjectSummary)
			continue
		}

		var subjects []models.SubjectSummary
		err := db.Model(&models.Subject{}).
			Select("uuid, subject_type, title, status, has_image").
			Where("subject_type = ?", subjectType).
			Order("id desc").
			Limit(5).
			Find(&subjects).Error
		if err != nil {
			return nil, err
		}
		results[subjectType] = subjects
		cache.SetCache(cacheKey, subjects)
	}

	return results, nil
}
