package models

type SubjectExportItem struct {
	UUID        string `json:"uuid"`
	SubjectType string `json:"subject_type"`
	Title       string `json:"title"`
	AltTitle    string `json:"alt_title"`
	PubDate     string `json:"pub_date"`
	Creator     string `json:"creator"`
	Press       string `json:"press"`
	Status      int    `json:"status"`
	Rating      int    `json:"rating"`
	Summary     string `json:"summary"`
	Comment     string `json:"comment"`
	ExternalURL string `json:"external_url"`
	MarkDate    string `json:"mark_date"`
	CreatedAt   string `json:"created_at"`
}

type SubjectExport struct {
	Subjects   []SubjectExportItem `json:"subjects"`
	ExportTime string              `json:"export_time"`
	TotalCount int                 `json:"total_count"`
}
