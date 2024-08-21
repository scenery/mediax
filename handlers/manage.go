package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/scenery/mediax/cache"
	"github.com/scenery/mediax/config"
	"github.com/scenery/mediax/database"
	"github.com/scenery/mediax/dataops"
	"github.com/scenery/mediax/helpers"
	"github.com/scenery/mediax/models"
)

func GetSubject(uuidStr string) (models.Subject, error) {
	cacheSubjectKey := fmt.Sprintf("subject:%s", uuidStr)
	cachedSubject, found := cache.GetCache(cacheSubjectKey)
	if found {
		return cachedSubject.(models.Subject), nil
	}

	var subject models.Subject
	db := database.GetDB()

	err := db.Table("subject").Where("uuid = ?", uuidStr).First(&subject).Error
	if err != nil {
		return subject, err
	}

	cache.SetCache(cacheSubjectKey, subject)
	return subject, nil
}

// 处理新增和编辑
func ManageSubject(w http.ResponseWriter, r *http.Request, uuidStr string) (string, error) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		log.Printf("failed to parse multipart form: %v", err)
		return "", err
	}

	data := make(map[string]string)
	for key := range r.Form {
		if key == "summary" || key == "comment" {
			data[key] = r.FormValue(key)
		} else {
			data[key] = strings.TrimSpace(r.FormValue(key))
		}
	}

	requiredFields := []string{"manage_type", "subject_type", "title", "status", "rating", "mark_date"}
	for _, field := range requiredFields {
		value, exists := data[field]
		if !exists || value == "" {
			return "", fmt.Errorf("missing required field: %s", field)
		}
	}

	subjectType := data["subject_type"]
	externalURL := data["external_url"]

	manageType, err := helpers.StringToInt(data["manage_type"])
	if err != nil {
		return "", fmt.Errorf("invalid manage type: %v", err)
	}

	if manageType == 3 || manageType == 4 {
		uuidStr = helpers.GenerateUUID()
	}

	hasImage := getHasImage(manageType, subjectType, uuidStr, externalURL)
	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		hasImage = 1
		imagePath := filepath.Join(config.ImageDir, subjectType, uuidStr+".jpg")
		err = dataops.SaveUploadedFile(file, imagePath)
		if err != nil {
			return "", err
		}
	}

	switch manageType {
	case 2:
		err = updateSubject(uuidStr, data, hasImage)
	case 3, 4:
		err = addSubject(uuidStr, data, hasImage)
		if err == nil && manageType == 4 && hasImage == 1 {
			dataops.MovePreDownloadedImage(subjectType, externalURL, uuidStr)
		}
	}
	if err != nil {
		return "", err
	}

	subjectURL := fmt.Sprintf("/%s/%s", subjectType, uuidStr)
	return subjectURL, nil
}

func updateSubject(uuidStr string, data map[string]string, hasImage int) error {
	db := database.GetDB()
	var subject models.Subject
	var err error

	err = db.Table("subject").Where("uuid = ?", uuidStr).Take(&subject).Error
	if err != nil {
		return err
	}

	subjectTypeOld := subject.SubjectType
	subjectStatusOld := subject.Status

	subjectTypeNew := data["subject_type"]
	subjectStatusNew, err := helpers.StringToInt(data["status"])
	if err != nil {
		return err
	}

	if hasImage == 0 {
		imagePath := filepath.Join(config.ImageDir, subjectTypeOld, uuidStr+".jpg")
		if _, err := os.Stat(imagePath); err == nil {
			hasImage = 1
		}
	}

	subject.SubjectType = subjectTypeNew
	subject.Title = data["title"]
	subject.AltTitle = data["alt_title"]
	subject.Creator = data["creator"]
	subject.Press = data["press"]
	subject.Status = subjectStatusNew
	subject.Rating, err = helpers.StringToInt(data["rating"])
	if err != nil {
		return err
	}
	subject.ExternalURL = data["external_url"]
	subject.HasImage = hasImage
	subject.Summary = data["summary"]
	subject.Comment = data["comment"]
	subject.PubDate = data["pub_date"]
	subject.MarkDate = data["mark_date"]
	subject.UpdatedAt = helpers.GetTimestamp()

	cache.DeleteCache(fmt.Sprintf("subject:%s", uuidStr))
	cache.DeleteSinglePageCache(subjectTypeOld, uuidStr, subjectStatusNew)

	if subjectStatusOld != subjectStatusNew {
		cache.ClearCommonCache(subjectTypeOld)
		cache.DeleteSinglePageCache(subjectTypeOld, uuidStr, subjectStatusOld)
	}

	if subjectTypeOld != subjectTypeNew {
		dataops.MoveDownloadedImage(subjectTypeOld, subjectTypeNew, uuidStr)

		cache.ClearCommonCache(subjectTypeOld)
		cache.ClearCommonCache(subjectTypeNew)
		cache.DeleteAfterPageCache(subjectTypeOld, uuidStr, subjectStatusOld)
		cache.ClearPagesCache(subjectTypeNew, subjectStatusNew)
	}
	return db.Save(&subject).Error
}

func addSubject(uuidStr string, data map[string]string, hasImage int) error {
	db := database.GetDB()
	var subject models.Subject
	var err error

	subjectStatus, err := helpers.StringToInt(data["status"])
	if err != nil {
		return err
	}
	subjectType := data["subject_type"]
	subject.UUID = uuidStr
	subject.SubjectType = subjectType
	subject.Title = data["title"]
	subject.AltTitle = data["alt_title"]
	subject.Creator = data["creator"]
	subject.Press = data["press"]
	subject.Status = subjectStatus
	subject.Rating, err = helpers.StringToInt(data["rating"])
	if err != nil {
		return err
	}
	subject.ExternalURL = data["external_url"]
	subject.HasImage = hasImage
	subject.Summary = data["summary"]
	subject.Comment = data["comment"]
	subject.PubDate = data["pub_date"]
	subject.MarkDate = data["mark_date"]
	subject.CreatedAt = helpers.GetTimestamp()
	subject.UpdatedAt = helpers.GetTimestamp()

	err = db.Create(&subject).Error
	if err != nil {
		return err
	}

	cache.ClearCommonCache(subjectType)
	cache.ClearPagesCache(subjectType, subjectStatus)
	return nil
}

func ManageDelSubject(uuidStr, subjectType string) error {
	db := database.GetDB()

	var subject models.Subject
	err := db.Select("status").Where("uuid = ?", uuidStr).First(&subject).Error
	if err != nil {
		return err
	}
	status := subject.Status

	err = db.Where("uuid = ?", uuidStr).Delete(&subject).Error
	if err != nil {
		return err
	}

	imageFilePath := filepath.Join(config.ImageDir, subjectType, uuidStr+".jpg")
	err = dataops.DeleteImage(imageFilePath)
	if err != nil {
		return err
	}

	cache.DeleteCache(fmt.Sprintf("subject:%s", uuidStr))
	cache.ClearCommonCache(subjectType)
	cache.DeleteAfterPageCache(subjectType, uuidStr, status)
	return nil
}

func getHasImage(manageType int, subjectType, uuidStr, externalURL string) int {
	var imagePath string

	if manageType == 2 {
		imagePath = filepath.Join(config.ImageDir, subjectType, uuidStr+".jpg")
		if _, err := os.Stat(imagePath); err == nil {
			return 1
		}
	}

	if manageType == 4 {
		imageName, err := dataops.PreDownloadImageName(externalURL)
		if err != nil {
			return 0
		}
		imageFilePath := filepath.Join(config.ImageDir, "temp", imageName)
		if _, err := os.Stat(imageFilePath); err == nil {
			return 1
		}
	}

	return 0
}

func CheckSubjectExistence(externalURL string) error {
	db := database.GetDB()
	var existUUID []string
	db.Model(&models.Subject{}).
		Where("external_url = ?", externalURL).
		Pluck("uuid", &existUUID)
	if len(existUUID) > 0 {
		return fmt.Errorf("subject with external URL <%s> already exists: %s", externalURL, existUUID[0])
	}
	return nil
}
