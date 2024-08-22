package models

import "html/template"

type SubjectView struct {
	PageTitle       string
	ManageType      int
	CreatorLabel    string
	PressLabel      string
	PubDateLabel    string
	SummaryLabel    string
	StatusText      string
	RatingStar      int
	ImageURL        string
	ExternalURLIcon template.HTML
	Subject         Subject
}

type CategoryViewItem struct {
	SubjectType  string
	SubjectURL   string
	Title        string
	AltTitle     string
	Creator      string
	Press        string
	PubDate      string
	MarkDate     string
	Rating       int
	StatusText   string
	CreatorLabel string
	PressLabel   string
	PubDateLabel string
	ImageURL     string
}

type StatusCounts struct {
	All     int64
	Todo    int64
	Doing   int64
	Done    int64
	OnHold  int64
	Dropped int64
}

type CategoryView struct {
	PageTitle    string
	Category     string
	Status       int
	StatusCounts StatusCounts
	CurrentPage  int
	TotalPages   int
	Subjects     []CategoryViewItem
}

type SearchView struct {
	PageTitle   string
	Query       string
	QueryType   string
	TotalCount  int64
	CurrentPage int
	TotalPages  int
	Subjects    []CategoryViewItem
}

type HomeViewItem struct {
	SubjectURL string
	Title      string
	MarkDate   string
	IsDoing    bool
	ImageURL   string
}

type HomeView struct {
	PageTitle    string
	FewBooks     bool
	FewMovies    bool
	FewTVs       bool
	FewAnimes    bool
	FewGames     bool
	RecentBooks  []HomeViewItem
	RecentMovies []HomeViewItem
	RecentTVs    []HomeViewItem
	RecentAnimes []HomeViewItem
	RecentGames  []HomeViewItem
}
