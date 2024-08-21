package handlers

import (
	"strings"

	"github.com/scenery/mediax/database"
	"github.com/scenery/mediax/models"
)

func GetSearchResult(query string) ([]models.SubjectSummary, int64, error) {
	db := database.GetDB()
	var subjects []models.SubjectSummary
	var total int64

	query = "%" + strings.TrimSpace(query) + "%"
	err := db.Table("subject").
		Where("title LIKE ?", query).
		Find(&subjects).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return subjects, total, nil
}
