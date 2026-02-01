package dataops

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	_ "golang.org/x/image/webp"

	"github.com/scenery/mediax/config"
	"github.com/scenery/mediax/database"
	"golang.org/x/image/draw"
)

func SaveRemoteImage(imageURL, imageFilePath string, interval bool) error {
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for URL %s: %v", imageURL, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:140.0) Gecko/20100101 Firefox/140.0")
	req.Header.Set("Accept", "image/avif,image/webp,image/png,image/svg+xml,image/*;q=0.8,*/*;q=0.5")
	req.Header.Set("Accept-Language", "zh,en-US;q=0.7,en;q=0.3")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download image from URL %s: %v", imageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("failed to download image from URL %s, status code: %d", imageURL, resp.StatusCode)
	}

	file, err := os.Create(imageFilePath)
	if err != nil {
		io.Copy(io.Discard, resp.Body)
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

func SaveUploadedImage(file io.Reader, subjectType, uuidStr string) error {
	targetFileName := uuidStr + ".jpg"
	targetDir := filepath.Join(config.ImageDir, subjectType)
	targetPath := filepath.Join(targetDir, targetFileName)

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("decoding uploaded image: %w", err)
	}

	outFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("creating target file %s: %w", targetPath, err)
	}
	defer outFile.Close()

	jpegOptions := &jpeg.Options{Quality: 90}
	err = jpeg.Encode(outFile, img, jpegOptions)
	if err != nil {
		os.Remove(targetPath)
		return fmt.Errorf("encoding image to JPEG and saving to %s: %w", targetPath, err)
	}

	err = GenerateThumbnail(subjectType, uuidStr, true)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	return nil
}

func DeleteImage(subjectType, uuidStr string) error {
	var err error

	imagePath := filepath.Join(config.ImageDir, subjectType, uuidStr+".jpg")
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil
	}
	err = os.Remove(imagePath)
	if err != nil {
		return err
	}

	thumbnailPath := filepath.Join(config.ThumbnailDir, subjectType, uuidStr+".jpg")
	if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
		return nil
	}
	err = os.Remove(thumbnailPath)
	if err != nil {
		return err
	}

	return nil
}

func MoveImage(subjectTypeOld, subjectTypeNew, uuidStr string) {
	var err error
	fileName := fmt.Sprintf("%s.jpg", uuidStr)

	err = moveFile(config.ImageDir, subjectTypeOld, subjectTypeNew, fileName)
	if err != nil {
		log.Printf("Error moving original image for UUID %s: %v", uuidStr, err)
		return
	}

	err = moveFile(config.ThumbnailDir, subjectTypeOld, subjectTypeNew, fileName)
	if err != nil {
		log.Printf("Error moving thumbnail image for UUID %s: %v", uuidStr, err)
		return
	}
}

func moveFile(baseDir, subjectTypeOld, subjectTypeNew, fileName string) error {
	sourcePath := filepath.Join(baseDir, subjectTypeOld, fileName)

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", sourcePath)
	} else if err != nil {
		return fmt.Errorf("failed to stat source file %s: %w", sourcePath, err)
	}

	destDir := filepath.Join(baseDir, subjectTypeNew)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}

	destPath := filepath.Join(destDir, fileName)

	if err := os.Rename(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to rename/move file from %s to %s: %w", sourcePath, destPath, err)
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
	err = SaveRemoteImage(imageURL, imageFilePath, false)
	if err != nil {
		log.Print(err)
		return
	}
}

func PreDownloadImageName(externalURL string) (string, error) {
	var imageName string
	pattern := regexp.MustCompile(`^https://(?:book|movie|www)?\.?(douban|bgm|bangumi)\.(?:com|tv)/(?:game|subject)/(\d+)/?$`)
	matched := pattern.MatchString(externalURL)
	if !matched {
		return imageName, errors.New("failed to get image name: unknown link source")
	}
	matches := pattern.FindStringSubmatch(externalURL)
	subjectType, subjectID := matches[1], matches[2]

	imageName = fmt.Sprintf("%s-%s.jpg", subjectType, subjectID)
	return imageName, nil
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

	err = GenerateThumbnail(subjectType, uuidStr, true)
	if err != nil {
		log.Printf("Error: %v", err)
	}
}

func GenerateThumbnail(subjectType, uuid string, force bool) error {
	originalFileName := uuid + ".jpg"
	originalPath := filepath.Join(config.ImageDir, subjectType, originalFileName)

	thumbnailFileName := uuid + ".jpg"
	thumbnailDir := filepath.Join(config.ThumbnailDir, subjectType)
	thumbnailPath := filepath.Join(thumbnailDir, thumbnailFileName)

	if !force {
		if _, err := os.Stat(thumbnailPath); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("checking thumbnail existence %s: %w", thumbnailPath, err)
		}
	}

	originalFile, err := os.Open(originalPath)
	if err != nil {
		return fmt.Errorf("opening original image %s: %w", originalPath, err)
	}
	defer originalFile.Close()

	img, _, err := image.Decode(originalFile)
	if err != nil {
		return fmt.Errorf("decoding image %s: %w", originalPath, err)
	}

	originalHeight := img.Bounds().Dy()
	var targetImg image.Image

	if originalHeight > config.ThumbnailHeight {
		originalWidth := img.Bounds().Dx()
		newWidth := int(float64(originalWidth) * (float64(config.ThumbnailHeight) / float64(originalHeight)))

		dst := image.NewRGBA(image.Rect(0, 0, newWidth, config.ThumbnailHeight))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
		targetImg = dst
	} else {
		targetImg = img
	}

	if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
		return fmt.Errorf("creating thumbnail directory %s: %w", thumbnailDir, err)
	}

	thumbnailFile, err := os.Create(thumbnailPath)
	if err != nil {
		return fmt.Errorf("creating thumbnail file %s: %w", thumbnailPath, err)
	}
	defer thumbnailFile.Close()

	err = jpeg.Encode(thumbnailFile, targetImg, &jpeg.Options{Quality: config.ThumbnailQuality})
	if err != nil {
		os.Remove(thumbnailPath)
		return fmt.Errorf("encoding thumbnail %s: %w", thumbnailPath, err)
	}

	return nil
}

func GenerateThumbnailFlag() {
	db := database.GetDB()

	type subject struct {
		UUID        string
		SubjectType string
	}

	var subjects []subject
	result := db.Table("subject").
		Select("uuid", "subject_type").
		Where("has_image = ?", 1).
		Find(&subjects)

	if result.Error != nil {
		log.Fatalf("Failed to query subjects with images: %v", result.Error)
	}

	log.Printf("Found %d subjects with images. Generating thumbnails...", len(subjects))

	generatedCount := 0
	errorCount := 0

	for _, subject := range subjects {
		err := GenerateThumbnail(subject.SubjectType, subject.UUID, false)
		if err != nil {
			log.Printf("Error generating thumbnail for %s/%s: %v", subject.SubjectType, subject.UUID, err)
			errorCount++
		} else {
			generatedCount++
			if generatedCount%100 == 0 {
				log.Printf("Processed %d subjects...", generatedCount)
			}
		}
	}

	log.Printf("Thumbnail generation finished. Total processed: %d, Errors: %d.", generatedCount, errorCount)
}
