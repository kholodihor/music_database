package models

type Album struct {
	Name  string `json:"name"`
	Year  string `json:"year"`
	Genre string `json:"genre"`
}

type Band struct {
	Name    string  `json:"name"`
	Country string  `json:"country"`
	Year    string  `json:"year"`
	Genre   string  `json:"genre"`
	Albums  []Album `json:"albums"`
}
