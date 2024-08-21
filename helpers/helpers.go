package helpers

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func GetTimestamp() int64 {
	return time.Now().Unix()
}

func StringToInt(value string) (int, error) {
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func GetSubjectType(subjectType string) string {
	switch subjectType {
	case "book":
		subjectType = "图书"
	case "movie":
		subjectType = "电影"
	case "tv":
		subjectType = "剧集"
	case "anime":
		subjectType = "番剧"
	case "game":
		subjectType = "游戏"
	default:
		subjectType = "未知"
	}
	return subjectType
}
