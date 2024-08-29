package config

const (
	// HTTP
	HTTP_PORT = 8080

	// Image
	ImageDir = "images"

	// Page
	PageSize = 10

	// Cache
	MaxCacheSubjects = 1000
	MaxCachePages    = 50
)

// API Config
var (
	CORS_HOST    = "*"
	RequestLimit = 50
)
