package model

type URL struct {
	ID        int64  `json:"-"`
	Code      string `json:"code"`
	Original  string `json:"original"`
	CreatedAt string `json:"created_at"`
	HitCount  int64  `json:"hit_count"`
}

type VisitLog struct {
	ID         int64  `json:"id"`
	Code       string `json:"code"`
	VisitedAt  string `json:"visited_at"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	Referer    string `json:"referer"`
	AcceptLang string `json:"accept_lang"`
	Origin     string `json:"origin"`
	Host       string `json:"host"`
}
