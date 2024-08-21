package dataops

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/scenery/mediax/database"
	"github.com/scenery/mediax/models"
)

func ExportToJSON(subjectType string, limit int) error {
	db := database.GetDB()

	var subjects []models.SubjectExportItem
	query := db.Model(&models.Subject{}).
		Select("uuid, subject_type, title, alt_title, pub_date, creator, press, status, rating, summary, comment, external_url, mark_date, created_at").
		Order("id DESC")
	if subjectType != "all" {
		query = query.Where("subject_type = ?", subjectType)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&subjects).Error; err != nil {
		return fmt.Errorf("failed to query subjects: %v", err)
	}

	exportData := models.SubjectExport{
		Subjects:   subjects,
		ExportTime: time.Now().Format(time.RFC3339),
		TotalCount: len(subjects),
	}

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	outputPath := fmt.Sprintf("mediaX_%s_%s.json", subjectType, time.Now().Format("20060102"))
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write JSON to file: %v", err)
	}

	fmt.Printf("Export complete! >> %s\n", outputPath)
	return nil
}
