package xrel

type Release struct {
	ID         string  `json:"id"`
	Dirname    string  `json:"dirname"`
	LinkHref   string  `json:"link_href"`
	Time       int     `json:"time"`
	GroupName  string  `json:"group_name"`
	Size       Size    `json:"size"`
	VideoType  string  `json:"video_type"`
	AudioType  string  `json:"audio_type"`
	NumRatings int     `json:"num_ratings"`
	ExtInfo    ExtInfo `json:"ext_info"`
	Comments   int     `json:"comments"`
}

type Size struct {
	Number int    `json:"number"`
	Unit   string `json:"unit"`
}

type ExtInfo struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	LinkHref   string   `json:"link_href"`
	Rating     float64  `json:"rating"`
	NumRatings int      `json:"num_ratings"`
	Uris       []string `json:"uris"`
}
