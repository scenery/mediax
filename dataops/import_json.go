package dataops

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/scenery/mediax/config"
	"github.com/scenery/mediax/database"
	"github.com/scenery/mediax/helpers"
	"github.com/scenery/mediax/models"
)

func ImportFromJSON(importType, filePath string, downloadImage bool) error {
	var err error

	switch importType {
	case "bangumi":
		err = importBangumiJSON(filePath, downloadImage)
		if err != nil {
			return err
		}
	case "douban":
		err = importDoubanJSON(filePath, downloadImage)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown import type: %s", importType)
	}
	fmt.Printf("Import complete! << %s\n", filePath)
	return nil
}

func importDoubanJSON(filePath string, downloadImage bool) error {
	db := database.GetDB()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var doubanData models.DoubanJson
	if err := json.Unmarshal(jsonData, &doubanData); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	existingRecords := getExistingRecords(db)
	existingImages := getExistingImages(config.ImageDir)

	total := len(doubanData.Interest)
	for index, item := range doubanData.Interest {
		var (
			altTitle string
			pubDate  string
			creator  string
			press    string
			summary  string
		)

		uuidStr := helpers.GenerateUUID()
		subjectType := item.Interest.Subject.Type
		externalURL := item.Interest.Subject.URL
		pubDate = getFirstPubdate(item.Interest.Subject.PubDate)

		if subjectType == "book" {
			if item.Interest.Subject.AltTitle != nil {
				altTitle = *item.Interest.Subject.AltTitle
			}
			if item.Interest.Subject.Author != nil {
				creator = joinStrings(*item.Interest.Subject.Author)
			}
			if item.Interest.Subject.Press != nil {
				press = joinStrings(*item.Interest.Subject.Press)
			}
		} else {
			if item.Interest.Subject.Directors != nil {
				creator = getCreator(*item.Interest.Subject.Directors)
			}
			tagParts := strings.Split(item.Interest.Subject.CardSubtitle, " / ")
			if len(tagParts) > 1 {
				press = tagParts[1]
			}
		}
		if item.Interest.Subject.Intro != nil {
			summary = *item.Interest.Subject.Intro
		}

		if existingSubject, exists := existingRecords[externalURL]; exists {
			if downloadImage {
				imageURL := item.Interest.Subject.Cover.Normal
				imageDir := filepath.Join(config.ImageDir, subjectType)
				if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
					return fmt.Errorf("failed to create images directory: %v", err)
				}
				imageFileName := fmt.Sprintf("%s.jpg", existingSubject.UUID)
				imageFilePath := filepath.Join(imageDir, imageFileName)
				relativeImagePath := filepath.Join(subjectType, imageFileName)

				if existingImages[relativeImagePath] {
					fmt.Printf("Image already exists at %s, skipping download...\n", imageFilePath)
				} else {
					if err := SaveImage(imageURL, imageFilePath, true); err != nil {
						fmt.Printf("Failed to download image: %v\n", err)
					} else {
						db.Model(&models.Subject{}).Where("external_url = ?", externalURL).Update("has_image", 1)
						fmt.Printf("Subject with external URL <%s> image updated: %s\n", externalURL, imageFilePath)
					}
				}
			}
			fmt.Printf("[%d/%d] Subject with external URL <%s> already exists, skipping...\n", index+1, total, externalURL)
			continue
		}

		hasImage := 1
		if !downloadImage {
			hasImage = 0
		}

		doubanSubject := models.Subject{
			UUID:        uuidStr,
			SubjectType: subjectType,
			Title:       item.Interest.Subject.Title,
			AltTitle:    altTitle,
			Creator:     creator,
			Press:       press,
			Status:      getDoubanStatus(item.Interest.Status),
			Rating:      item.Interest.Rating.Value * 2,
			HasImage:    hasImage,
			ExternalURL: externalURL,
			Summary:     summary,
			Comment:     item.Interest.Comment,
			PubDate:     pubDate,
			MarkDate:    parseDate(item.Interest.CreateTime),
			CreatedAt:   helpers.GetTimestamp(),
			UpdatedAt:   helpers.GetTimestamp(),
		}

		if err := db.Create(&doubanSubject).Error; err != nil {
			fmt.Printf("[%d/%d] Failed to insert subject record for %s: %v\n", index+1, total, externalURL, err)
		} else {
			fmt.Printf("[%d/%d] Subject with external URL <%s> inserted successfully\n", index+1, total, externalURL)
		}

		if downloadImage {
			imageURL := item.Interest.Subject.Cover.Normal
			imageDir := filepath.Join(config.ImageDir, subjectType)
			if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create images directory: %v", err)
			}
			imageFileName := fmt.Sprintf("%s.jpg", uuidStr)
			imageFilePath := filepath.Join(imageDir, imageFileName)
			if err := SaveImage(imageURL, imageFilePath, true); err != nil {
				fmt.Printf("Failed to download image: %v\n", err)
			} else {
				fmt.Printf("<%s> Cover image downloaded: %s\n", externalURL, imageFilePath)
			}
		}
	}
	return nil
}

func importBangumiJSON(filePath string, downloadImage bool) error {
	db := database.GetDB()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var bangumiData models.BangumiJson
	err = json.Unmarshal(byteValue, &bangumiData)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	existingRecords := getExistingRecords(db)
	existingImages := getExistingImages(config.ImageDir)

	total := len(bangumiData.Data)
	for i := total - 1; i >= 0; i-- {
		item := bangumiData.Data[i]
		uuidStr := helpers.GenerateUUID()
		subjectType := getBangumiType(item.SubjectType)
		if subjectType == "tv" {
			eps := item.Subject.Eps
			if eps == 1 {
				subjectType = "movie"
			}
		}

		externalURL := fmt.Sprintf("https://bgm.tv/subject/%d", item.Subject.ID)
		if existingSubject, exists := existingRecords[externalURL]; exists {
			if downloadImage {
				imageURL := item.Subject.Images.Common
				imageDir := filepath.Join(config.ImageDir, subjectType)
				if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
					return fmt.Errorf("failed to create images directory: %v", err)
				}
				imageFileName := fmt.Sprintf("%s.jpg", existingSubject.UUID)
				imageFilePath := filepath.Join(imageDir, imageFileName)
				relativeImagePath := filepath.Join(subjectType, imageFileName)

				if existingImages[relativeImagePath] {
					fmt.Printf("Image already exists at %s, skipping download...\n", imageFilePath)
				} else {
					if err := SaveImage(imageURL, imageFilePath, true); err != nil {
						fmt.Printf("Failed to download image: %v\n", err)
					} else {
						db.Model(&models.Subject{}).Where("external_url = ?", externalURL).Update("has_image", 1)
						fmt.Printf("Subject with external URL <%s> image updated: %s\n", externalURL, imageFilePath)
					}
				}
			}
			fmt.Printf("[%d/%d] Subject with external URL <%s> already exists, skipping...\n", total-i, total, externalURL)
			continue
		}

		hasImage := 1
		if !downloadImage {
			hasImage = 0
		}
		title, altTitle := getBangumiTitle(item.Subject.Name, item.Subject.NameCN)

		bangumiSubject := models.Subject{
			UUID:        uuidStr,
			SubjectType: subjectType,
			Title:       title,
			AltTitle:    altTitle,
			Status:      getBangumiStatus(item.Type),
			Rating:      item.Rate,
			ExternalURL: externalURL,
			HasImage:    hasImage,
			Summary:     item.Subject.ShortSummary,
			Comment:     item.Comment,
			PubDate:     item.Subject.Date,
			MarkDate:    parseDate(item.UpdatedAt),
			CreatedAt:   helpers.GetTimestamp(),
			UpdatedAt:   helpers.GetTimestamp(),
		}

		if err := db.Create(&bangumiSubject).Error; err != nil {
			fmt.Printf("[%d/%d] Failed to insert subject record for %s: %v\n", total-i, total, externalURL, err)
		} else {
			fmt.Printf("[%d/%d] Subject with external URL <%s> inserted successfully\n", total-i, total, externalURL)
		}

		if downloadImage {
			imageURL := item.Subject.Images.Common
			imageDir := filepath.Join(config.ImageDir, subjectType)
			if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create images directory: %v", err)
			}
			imageFileName := fmt.Sprintf("%s.jpg", uuidStr)
			imageFilePath := filepath.Join(imageDir, imageFileName)
			if err := SaveImage(imageURL, imageFilePath, true); err != nil {
				fmt.Printf("Failed to download image: %v\n", err)
			} else {
				fmt.Printf("<%s> Cover image downloaded: %s\n", externalURL, imageFilePath)
			}
		}
	}
	return nil
}
