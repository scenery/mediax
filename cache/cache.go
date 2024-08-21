package cache

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/scenery/mediax/helpers"
	"github.com/scenery/mediax/models"
)

// 首页缓存: home:{subject_type}
// 分类页面缓存: page:{subject_type}:{status}:{page_number}
// 分类总数缓存: count:{subject_type}
// 条目缓存: subject:{uuid}
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

var (
	cacheItem    sync.Map // 首页、条目、分类总数缓存
	cachePages   sync.Map // 分类页面缓存
	pageSubjects sync.Map // 分类某页中所有条目的 UUID
)

// 获取缓存项
func GetCache(key string) (interface{}, bool) {
	cacheType := getCacheType(key)

	switch cacheType {
	case "common":
		if value, found := cacheItem.Load(key); found {
			item := value.(CacheItem)
			if time.Now().After(item.Expiration) {
				cacheItem.Delete(key)
				return nil, false
			}
			// fmt.Printf("已获得缓存: %s\n", key)
			return item.Value, true
		}
	case "page":
		if value, found := cachePages.Load(key); found {
			item := value.(CacheItem)
			if time.Now().After(item.Expiration) {
				cachePages.Delete(key)
				return nil, false
			}
			// fmt.Printf("已获得缓存: %s\n", key)
			return item.Value, true
		}
	}

	return nil, false
}

// 设置缓存项
func SetCache(key string, value interface{}, duration ...time.Duration) {
	var expiration time.Time
	if len(duration) > 0 {
		expiration = time.Now().Add(duration[0])
	} else {
		expiration = time.Now().Add(10 * 365 * 24 * time.Hour) // 默认 10 年
	}

	item := CacheItem{
		Value:      value,
		Expiration: expiration,
	}

	cacheType := getCacheType(key)

	switch cacheType {
	case "common":
		// fmt.Printf("已设置缓存: %s\n", key)
		cacheItem.Store(key, item)
	case "page":
		// fmt.Printf("已设置缓存: %s\n", key)
		cachePages.Store(key, item)
		updatePageSubjects(key, value)
	}
}

// 删除缓存项
func DeleteCache(key string) {
	cacheType := getCacheType(key)

	switch cacheType {
	case "common":
		// fmt.Printf("已删除缓存: %s\n", key)
		cacheItem.Delete(key)
	case "page":
		// fmt.Printf("已删除缓存: %s\n", key)
		cachePages.Delete(key)
	}
}

// 清理通用缓存
func ClearCommonCache(subjectType string) {
	homeCacheKey := fmt.Sprintf("home:%s", subjectType)
	DeleteCache(homeCacheKey)

	countCacheKey := fmt.Sprintf("count:%s", subjectType)
	DeleteCache(countCacheKey)
}

// 清理新增操作时该分类下所有页缓存
func ClearPagesCache(subjectType string, status int) {
	statusPrefix := fmt.Sprintf("page:%s:%d:", subjectType, status)
	allStatusPrefix := fmt.Sprintf("page:%s:0:", subjectType)

	cachePages.Range(func(key, value interface{}) bool {
		cacheKey := key.(string)
		if strings.HasPrefix(cacheKey, statusPrefix) || strings.HasPrefix(cacheKey, allStatusPrefix) {
			DeleteCache(cacheKey)
			deletePageSubjects(subjectType, status, 0)
			deletePageSubjects(subjectType, 0, 0)
		}
		return true
	})
}

// 清理更新操作时条目所在页缓存
func DeleteSinglePageCache(subjectType, uuidStr string, status int) {
	var subjectPage int

	pageSubjects.Range(func(key, value interface{}) bool {
		uuids := value.([]string)
		for _, uuid := range uuids {
			if uuid == uuidStr {
				subjectPage = getPageNumber(key.(string))
				return false
			}
		}
		return true
	})

	if subjectPage == 0 {
		return
	}

	cacheKey := fmt.Sprintf("page:%s:%d:%d", subjectType, status, subjectPage)
	DeleteCache(cacheKey)
	deletePageSubjects(subjectType, status, subjectPage)

	allStatusCacheKey := fmt.Sprintf("page:%s:0:%d", subjectType, subjectPage)
	DeleteCache(allStatusCacheKey)
	deletePageSubjects(subjectType, 0, subjectPage)
}

// 清理删除操作时条目所在页及之后所有页的缓存
func DeleteAfterPageCache(subjectType, uuidStr string, status int) {
	var subjectPage int

	pageSubjects.Range(func(key, value interface{}) bool {
		uuids := value.([]string)
		for _, uuid := range uuids {
			if uuid == uuidStr {
				subjectPage = getPageNumber(key.(string))
				return false
			}
		}
		return true
	})

	if subjectPage == 0 {
		return
	}

	cachePages.Range(func(key, value interface{}) bool {
		cacheKey := key.(string)
		parts := strings.Split(cacheKey, ":")
		if len(parts) != 4 {
			return true
		}

		cacheSubjectType := parts[1]
		cacheStatus, _ := helpers.StringToInt(parts[2])
		cachedPageNumber := getPageNumber(cacheKey)

		if cacheSubjectType != subjectType {
			return true
		}

		if (cacheStatus == status || cacheStatus == 0) && cachedPageNumber >= subjectPage {
			DeleteCache(cacheKey)
			deletePageSubjects(subjectType, cacheStatus, cachedPageNumber)
		}

		return true
	})
}

// 更新页码到 UUID 的映射
func updatePageSubjects(key string, value interface{}) {
	if subjects, ok := value.([]models.SubjectSummary); ok {
		uuids := make([]string, len(subjects))
		for i, subject := range subjects {
			uuids[i] = subject.UUID
		}
		pageNumber := getPageNumber(key)
		subjectType := strings.SplitN(key, ":", 4)[1]
		status := strings.SplitN(key, ":", 4)[2]
		pageSubjects.Store(fmt.Sprintf("page:%s:%s:%d", subjectType, status, pageNumber), uuids)
		// fmt.Printf("已设置页码-UUID映射缓存: page:%s:%s:%d\n", subjectType, status, pageNumber)
	}
}

// 删除页码到 UUID 的映射
func deletePageSubjects(subjectType string, status int, pageNumber int) {
	if pageNumber == 0 {
		pageSubjects.Range(func(key, value interface{}) bool {
			keyName := key.(string)
			if strings.HasPrefix(keyName, fmt.Sprintf("page:%s:%d:", subjectType, status)) {
				// fmt.Printf("已删除页码-UUID映射缓存: %s\n", keyName)
				pageSubjects.Delete(keyName)
			}
			return true
		})
	} else {
		cacheKey := fmt.Sprintf("page:%s:%d:%d", subjectType, status, pageNumber)
		pageSubjects.Delete(cacheKey)
		// fmt.Printf("已删除页码-UUID映射缓存: %s\n", cacheKey)
	}
}

// 从缓存键中提取页码
func getPageNumber(key string) int {
	parts := strings.Split(key, ":")
	if len(parts) < 4 {
		return 0
	}
	pageNumber, _ := helpers.StringToInt(parts[3])
	return pageNumber
}

// 判断缓存类型
func getCacheType(key string) string {
	if strings.HasPrefix(key, "page:") {
		return "page"
	}
	return "common"
}
