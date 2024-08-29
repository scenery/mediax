package routes

import (
	"fmt"
	"net/http"

	"github.com/scenery/mediax/config"
	"github.com/scenery/mediax/dataops"
	"github.com/scenery/mediax/helpers"
)

func handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Origin", config.CORS_HOST)

	switch r.Method {
	case http.MethodGet:
		query := r.URL.Query()
		subjectType := query.Get("type")
		queryLimit := query.Get("limit")
		queryOffset := query.Get("offset")

		validTypes := map[string]bool{
			"all":   true,
			"book":  true,
			"movie": true,
			"tv":    true,
			"anime": true,
			"game":  true,
		}
		if subjectType == "" {
			subjectType = "all"
		} else if !validTypes[subjectType] {
			handleAPIError(w, http.StatusBadRequest, "invalid subject type")
			return
		}

		limit := config.RequestLimit
		if queryLimit != "" {
			var err error
			limit, err = helpers.StringToInt(queryLimit)
			if err != nil {
				handleAPIError(w, http.StatusBadRequest, "invalid limit")
				return
			}
			if limit < 1 || limit > config.RequestLimit {
				limit = config.RequestLimit
			}
		}

		offset := 0
		if queryOffset != "" {
			var err error
			offset, err = helpers.StringToInt(queryOffset)
			if err != nil {
				handleAPIError(w, http.StatusBadRequest, "invalid offset")
				return
			}
			if offset < 1 {
				offset = 0
			}
		}

		responseJSON, err := dataops.ExportToJSONAPI(subjectType, limit, offset)
		if err != nil {
			handleAPIError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Write(responseJSON)
		return
	default:
		handleAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
}

func handleAPIError(w http.ResponseWriter, errStatus int, errMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errStatus)
	response := fmt.Sprintf(`{"error": "%s"}`, errMessage)
	w.Write([]byte(response))
}
