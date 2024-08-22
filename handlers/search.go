package handlers

import (
	"fmt"
	"time"

	"github.com/scenery/mediax/cache"
	"github.com/scenery/mediax/database"
	"github.com/scenery/mediax/models"
)

func GetSearchResult(query string, page int, pageSize int) ([]models.SubjectSummary, int64, error) {
	db := database.GetDB()
	var subjects []models.SubjectSummary
	var total int64
	likeQuery := "%" + query + "%"

	cacheSearchKey := fmt.Sprintf("search:%s:%d", query, page)
	cacheCountKey := fmt.Sprintf("count:search:%s", query)

	if cachedCounts, found := cache.GetCache(cacheCountKey); found {
		total = cachedCounts.(int64)
	} else {
		err := db.Table("subject").
			Where("title LIKE ?", likeQuery).
			Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		cache.SetCache(cacheCountKey, total, 10*time.Minute)
	}

	if cachedSubjects, found := cache.GetCache(cacheSearchKey); found {
		subjects = cachedSubjects.([]models.SubjectSummary)
	} else {
		offset := (page - 1) * pageSize
		err := db.Table("subject").
			Where("title LIKE ?", likeQuery).
			Offset(offset).
			Limit(pageSize).
			Find(&subjects).Error
		if err != nil {
			return nil, 0, err
		}
		cache.SetCache(cacheSearchKey, subjects, 10*time.Minute)
	}

	return subjects, total, nil
}
