package models

// Douban
type DoubanCover struct {
	Normal string `json:"normal"`
}

type DoubanDirector struct {
	Name string `json:"name"`
}

// import from file
type DoubanSubject struct {
	Title        string            `json:"title"`
	AltTitle     *string           `json:"book_subtitle"`
	URL          string            `json:"url"`
	PubDate      []string          `json:"pubdate"`
	Directors    *[]DoubanDirector `json:"directors"`
	Author       *[]string         `json:"author"`
	Press        *[]string         `json:"press"`
	CardSubtitle string            `json:"card_subtitle"`
	Intro        *string           `json:"intro"`
	Type         string            `json:"type"`
	Cover        DoubanCover       `json:"pic"`
}

type DoubanDetail struct {
	Comment string `json:"comment"`
	Rating  struct {
		Value int `json:"value"`
	} `json:"rating"`
	Status     string        `json:"status"`
	CreateTime string        `json:"create_time"`
	Subject    DoubanSubject `json:"subject"`
}

type DoubanItem struct {
	Type     string       `json:"type"`
	Status   string       `json:"status"`
	Interest DoubanDetail `json:"interest"`
}

type DoubanJson struct {
	Interest []DoubanItem `json:"interest"`
}

// import from api
type DoubanBookSubject struct {
	Title        string      `json:"title"`
	AltTitle     string      `json:"book_subtitle"`
	PubDate      []string    `json:"pubdate"`
	Author       []string    `json:"author"`
	Press        []string    `json:"press"`
	CardSubtitle string      `json:"card_subtitle"`
	Intro        string      `json:"intro"`
	Type         string      `json:"type"`
	Cover        DoubanCover `json:"pic"`
}

type DoubanMovieSubject struct {
	Title        string           `json:"title"`
	AltTitle     string           `json:"original_title"`
	PubDate      []string         `json:"pubdate"`
	Directors    []DoubanDirector `json:"directors"`
	CardSubtitle string           `json:"card_subtitle"`
	Intro        string           `json:"intro"`
	Type         string           `json:"type"`
	Cover        DoubanCover      `json:"pic"`
}

type DoubanGameSubject struct {
	Title       string      `json:"title"`
	TitleCN     string      `json:"cn_name"`
	ReleaseDate string      `json:"release_date"`
	Developer   []string    `json:"developers"`
	Publisher   []string    `json:"publishers"`
	Intro       string      `json:"intro"`
	Type        string      `json:"type"`
	Cover       DoubanCover `json:"pic"`
}

// Bangumi
type BangumiImage struct {
	Common string `json:"common"`
}

// import from file
type BangumiSubject struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	NameCN       string       `json:"name_cn"`
	ShortSummary string       `json:"short_summary"`
	Date         string       `json:"date"`
	Eps          int          `json:"eps"`
	Images       BangumiImage `json:"images"`
}

type BangumiItem struct {
	UpdatedAt   string         `json:"updated_at"`
	Comment     string         `json:"comment"`
	Subject     BangumiSubject `json:"subject"`
	Rate        int            `json:"rate"`
	Type        int            `json:"type"`
	SubjectID   int            `json:"subject_id"`
	SubjectType int            `json:"subject_type"`
}

type BangumiJson struct {
	Data []BangumiItem `json:"data"`
}

// import from api
type Infobox struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type BangumiSubjectDetail struct {
	ID      int          `json:"id"`
	Type    int          `json:"type"`
	Name    string       `json:"name"`
	NameCN  string       `json:"name_cn"`
	Summary string       `json:"summary"`
	Date    string       `json:"date"`
	Images  BangumiImage `json:"images"`
	Infobox []Infobox    `json:"infobox"`
}
