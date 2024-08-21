package handlers

import (
	"fmt"

	"github.com/scenery/mediax/cache"
	"github.com/scenery/mediax/config"
	"github.com/scenery/mediax/database"
	"github.com/scenery/mediax/models"
)

func GetSubjectsByType(subjectType string, status int, page int, pageSize int) ([]models.SubjectSummary, error) {
	db := database.GetDB()
	var subjects []models.SubjectSummary

	cacheSubjectsKey := fmt.Sprintf("page:%s:%d:%d", subjectType, status, page)

	if page <= config.MaxCachePages {
		if cachedSubjects, found := cache.GetCache(cacheSubjectsKey); found {
			subjects = cachedSubjects.([]models.SubjectSummary)
			return subjects, nil
		}
	}

	query := db.Table("subject").
		Where("subject_type = ?", subjectType).
		Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	if status > 0 {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&subjects).Error
	if err != nil {
		return nil, err
	}

	if page <= config.MaxCachePages {
		cache.SetCache(cacheSubjectsKey, subjects)
	}

	return subjects, nil
}

func GetStatusCounts(subjectType string) (models.StatusCounts, error) {
	db := database.GetDB()
	var counts models.StatusCounts

	cacheKey := fmt.Sprintf("count:%s", subjectType)
	if cachedCounts, found := cache.GetCache(cacheKey); found {
		return cachedCounts.(models.StatusCounts), nil
	}

	rows, err := db.Table("subject").
		Select("status, COUNT(*) as count").
		Where("subject_type = ?", subjectType).
		Group("status").
		Rows()
	if err != nil {
		return counts, err
	}
	defer rows.Close()

	for rows.Next() {
		var status int
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return counts, err
		}
		switch status {
		case 1:
			counts.Todo = count
		case 2:
			counts.Doing = count
		case 3:
			counts.Done = count
		case 4:
			counts.OnHold = count
		case 5:
			counts.Dropped = count
		}
	}

	counts.All = counts.Todo + counts.Doing + counts.Done + counts.OnHold + counts.Dropped

	cache.SetCache(cacheKey, counts)

	return counts, nil
}
