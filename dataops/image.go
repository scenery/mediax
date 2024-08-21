package dataops

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/scenery/mediax/config"
)

func SaveImage(imageURL, imageFilePath string, interval bool) error {
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for URL %s: %v", imageURL, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:128.0) Gecko/20100101 Firefox/128.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download image from URL %s: %v", imageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image from URL %s, status code: %d", imageURL, resp.StatusCode)
	}

	file, err := os.Create(imageFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file for image %s: %v", imageURL, err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save image from URL %s: %v", imageURL, err)
	}

	if interval {
		time.Sleep(1 * time.Second)
	}

	return nil
}

func DeleteImage(imageFilePath string) error {
	if _, err := os.Stat(imageFilePath); os.IsNotExist(err) {
		return nil
	}

	err := os.Remove(imageFilePath)
	if err != nil {
		return err
	}

	return nil
}

func SaveUploadedFile(file multipart.File, path string) error {
	limitedFile := io.LimitReader(file, 8*1024*1024) // 8 MB

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, limitedFile)
	if err != nil {
		return err
	}

	return nil
}

func PreDownloadImage(imageURL, externalURL string) {
	imageName, err := PreDownloadImageName(externalURL)
	if err != nil {
		log.Print(err)
		return
	}

	imageDir := filepath.Join(config.ImageDir, "temp")
	err = os.MkdirAll(imageDir, os.ModePerm)
	if err != nil {
		log.Print(err)
		return
	}

	imageFilePath := filepath.Join(imageDir, imageName)
	err = SaveImage(imageURL, imageFilePath, false)
	if err != nil {
		log.Print(err)
		return
	}
}

func MoveDownloadedImage(subjectTypeOld, subjectTypeNew, uuidStr string) {
	var err error
	sourceFileName := fmt.Sprintf("%s.jpg", uuidStr)
	sourceFilePath := filepath.Join(config.ImageDir, subjectTypeOld, sourceFileName)
	if _, err = os.Stat(sourceFilePath); os.IsNotExist(err) {
		log.Print(err)
		return
	}

	destDir := filepath.Join(config.ImageDir, subjectTypeNew)
	err = os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		log.Print(err)
		return
	}

	destFileName := fmt.Sprintf("%s.jpg", uuidStr)
	destFilePath := filepath.Join(destDir, destFileName)

	err = os.Rename(sourceFilePath, destFilePath)
	if err != nil {
		log.Print(err)
		return
	}
}

func MovePreDownloadedImage(subjectType, externalURL, uuidStr string) {
	imageName, err := PreDownloadImageName(externalURL)
	if err != nil {
		log.Print(err)
		return
	}

	sourceFilePath := filepath.Join(config.ImageDir, "temp", imageName)
	if _, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
		log.Print(err)
		return
	}

	destDir := filepath.Join(config.ImageDir, subjectType)
	err = os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		log.Print(err)
		return
	}

	destFileName := fmt.Sprintf("%s.jpg", uuidStr)
	destFilePath := filepath.Join(destDir, destFileName)

	err = os.Rename(sourceFilePath, destFilePath)
	if err != nil {
		log.Print(err)
		return
	}
}

func PreDownloadImageName(externalURL string) (string, error) {
	var imageName string
	pattern := regexp.MustCompile(`^https://(?:book|movie)?\.?(douban|bgm|bangumi)\.(?:com|tv)/subject/(\d+)/?$`)
	matched := pattern.MatchString(externalURL)
	if !matched {
		return imageName, errors.New("failed to get image name: unknown link source")
	}
	matches := pattern.FindStringSubmatch(externalURL)
	subjectType, subjectID := matches[1], matches[2]

	imageName = fmt.Sprintf("%s-%s.jpg", subjectType, subjectID)
	return imageName, nil
}
