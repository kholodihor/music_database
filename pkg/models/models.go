package models

type Album struct {
	Name      string `json:"name"`
	Year      int    `json:"year"`
	Genre     string `json:"genre"`
	ImagePath string `json:"imagePath,omitempty"`
}

type Band struct {
	Name    string  `json:"name"`
	Genre   string  `json:"genre"`
	Country string  `json:"country"`
	Year    int     `json:"year"`
	Albums  []Album `json:"albums"`
}
