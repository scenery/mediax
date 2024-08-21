package dataops

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/scenery/mediax/models"
)

func FetchMediaInfo(subjectType, subjectID, apiTarget, externalURL string) (models.Subject, error) {
	const doubanAPIHost = "frodo.douban.com"
	const doubanAPIToken = "0ac44ae016490db2204ce0a042db2916"
	const bangumiAPIHost = "api.bgm.tv"
	// const bangumiAPIToken = ""
	var headers map[string]string
	var data models.Subject
	var imageURL string
	var err error
	var requestURL string

	switch apiTarget {
	case "douban":
		apiSubjectType := "movie"
		if subjectType == "book" {
			apiSubjectType = "book"
		}
		apiURL := fmt.Sprintf("https://%s/api/v2/%s/%s", doubanAPIHost, apiSubjectType, subjectID)
		params := fmt.Sprintf("?apiKey=%s", doubanAPIToken)
		requestURL = apiURL + params
		headers = map[string]string{
			"Host":       doubanAPIHost,
			"User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 15_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.16(0x18001023) NetType/WIFI Language/zh_CN",
			"Referer":    "https://servicewechat.com/wx2f9b06c1de1ccfca/84/page-frame.html",
		}
		data, imageURL, err = getDoubanData(subjectType, requestURL, headers)
		if err != nil {
			return data, err
		}
	case "bangumi":
		requestURL = fmt.Sprintf("https://%s/v0/subjects/%s", bangumiAPIHost, subjectID)
		headers = map[string]string{
			"Accept":     "application/json",
			"User-Agent": "Mozilla/5.0",
		}
		data, imageURL, err = getBangumiData(subjectType, requestURL, headers)
		if err != nil {
			return data, err
		}
	default:
		return data, fmt.Errorf("wrong subject type: %s", subjectType)
	}

	PreDownloadImage(imageURL, externalURL)

	return data, err
}

func getDoubanData(subjectType, requestURL string, headers map[string]string) (models.Subject, string, error) {
	subject := models.Subject{}
	var imageURL string
	var err error

	jsonData, err := requestAPI(requestURL, headers)
	if err != nil {
		return subject, imageURL, err
	}

	var doubanSubject interface{}
	if subjectType == "book" {
		doubanSubject = &models.DoubanBookSubject{}
	} else {
		doubanSubject = &models.DoubanMovieSubject{}
	}

	if err := json.Unmarshal(jsonData, doubanSubject); err != nil {
		return subject, imageURL, err
	}

	var (
		title    string
		altTitle string
		creator  string
		press    string
		pubDate  string
		summary  string
	)

	switch subject := doubanSubject.(type) {
	case *models.DoubanBookSubject:
		title = subject.Title
		altTitle = subject.AltTitle
		creator = joinStrings(subject.Author)
		press = joinStrings(subject.Press)
		pubDate = getFirstPubdate(subject.PubDate)
		summary = subject.Intro
		imageURL = subject.Cover.Normal
	case *models.DoubanMovieSubject:
		title = subject.Title
		altTitle = subject.AltTitle
		creator = getCreator(subject.Directors)
		tagParts := strings.Split(subject.CardSubtitle, " / ")
		if len(tagParts) > 1 {
			press = tagParts[1]
		} else {
			press = ""
		}
		pubDate = getFirstPubdate(subject.PubDate)
		summary = subject.Intro
		imageURL = subject.Cover.Normal
	}

	subject = models.Subject{
		Title:    title,
		AltTitle: altTitle,
		Creator:  creator,
		Press:    press,
		Summary:  summary,
		PubDate:  pubDate,
	}

	return subject, imageURL, err
}

func getBangumiData(subjectType, requestURL string, headers map[string]string) (models.Subject, string, error) {
	subject := models.Subject{}
	var imageURL string
	var err error

	jsonSubject, err := requestAPI(requestURL, headers)
	if err != nil {
		return subject, imageURL, err
	}

	var bangumiSubject models.BangumiSubjectDetail
	err = json.Unmarshal(jsonSubject, &bangumiSubject)
	if err != nil {
		fmt.Println(err)
		return subject, imageURL, err
	}

	var (
		title    string
		altTitle string
		creator  string
		press    string
		pubDate  string
		summary  string
	)

	switch subjectType {
	case "anime":
		creator, press = extractInfobox(bangumiSubject.Infobox, "导演", "製作")
	case "game":
		creator, press = extractInfobox(bangumiSubject.Infobox, "游戏开发商", "发行")
	case "book":
		creator, press = extractInfobox(bangumiSubject.Infobox, "作者", "出版社")
	default:
		creator, press = extractInfobox(bangumiSubject.Infobox, "导演", "国家/地区")
	}

	title, altTitle = getBangumiTitle(bangumiSubject.Name, bangumiSubject.NameCN)
	pubDate = bangumiSubject.Date
	summary = bangumiSubject.Summary
	imageURL = bangumiSubject.Images.Common

	subject = models.Subject{
		SubjectType: getBangumiType(bangumiSubject.Type),
		Title:       title,
		AltTitle:    altTitle,
		Creator:     creator,
		Press:       press,
		Summary:     summary,
		PubDate:     pubDate,
		ExternalURL: fmt.Sprintf("https://bgm.tv/subject/%d", bangumiSubject.ID),
	}

	return subject, imageURL, nil
}

func extractInfobox(infobox []models.Infobox, creatorKey, pressKey string) (string, string) {
	var creator, press string
	for _, item := range infobox {
		if creator != "" && press != "" {
			break
		}
		switch item.Key {
		case creatorKey:
			if value, ok := item.Value.(string); ok {
				creator = value
			}
		case pressKey:
			if value, ok := item.Value.(string); ok {
				if pressKey == "製作" && strings.Contains(value, "；") {
					press = strings.Split(value, "；")[0]
				} else {
					press = value
				}
			}
		}
	}
	return creator, press
}

func requestAPI(requestURL string, headers map[string]string) ([]byte, error) {
	const maxRetries = 2
	const retryDelay = 2 * time.Second
	client := &http.Client{Timeout: 10 * time.Second}
	var body []byte
	var err error

	for i := 0; i < maxRetries; i++ {
		req, reqErr := http.NewRequest("GET", requestURL, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, respErr := client.Do(req)
		if respErr != nil {
			err = respErr
			fmt.Printf("Attempt %d: failed to fetch data, error: %v\n", i+1, err)
			time.Sleep(retryDelay)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("bad status: %s", resp.Status)
			fmt.Printf("Attempt %d: received bad status, error: %v\n", i+1, err)
			time.Sleep(retryDelay)
			continue
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Attempt %d: failed to read response body, error: %v\n", i+1, err)
			time.Sleep(retryDelay)
			continue
		}
		return body, nil
	}
	return nil, fmt.Errorf("failed to fetch data after multiple attempts\n%s", err)
}
