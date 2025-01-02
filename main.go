package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"music-database/db"
	"music-database/models"
)

type Server struct {
	db        *db.Driver
	templates *template.Template
}

type PageData struct {
	Bands []models.Band
}

func formatBandName(name string) string {
	// Convert to lowercase and replace spaces with underscores
	return strings.ToLower(strings.ReplaceAll(name, " ", "_"))
}

func NewServer(database *db.Driver) *Server {
	funcMap := template.FuncMap{
		"formatBandName": formatBandName,
	}
	templates := template.New("").Funcs(funcMap)
	templates = template.Must(templates.ParseGlob("templates/*.html"))
	return &Server{
		db:        database,
		templates: templates,
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	results, err := s.db.Query("bands", db.Query{Field: "genre", Operator: "eq", Value: "Progressive Rock"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var bands []models.Band
	for _, result := range results {
		if data, ok := result.(map[string]interface{}); ok {
			band := models.Band{
				Name:    data["name"].(string),
				Country: data["country"].(string),
				Year:    fmt.Sprintf("%v", data["year"]),
				Genre:   data["genre"].(string),
			}
			
			if albumsData, ok := data["albums"].([]interface{}); ok {
				for _, albumData := range albumsData {
					if album, ok := albumData.(map[string]interface{}); ok {
						albumName := album["name"].(string)
						albumYear := fmt.Sprintf("%v", album["year"])
						albumGenre := album["genre"].(string)
						band.Albums = append(band.Albums, models.Album{
							Name:  albumName,
							Year:  albumYear,
							Genre: albumGenre,
						})
					}
				}
			}
			bands = append(bands, band)
		}
	}

	data := PageData{Bands: bands}
	s.templates.ExecuteTemplate(w, "layout.html", data)
}

func (s *Server) handleAddBand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	band := models.Band{
		Name:    r.FormValue("name"),
		Country: r.FormValue("country"),
		Year:    r.FormValue("year"),
		Genre:   r.FormValue("genre"),
		Albums:  []models.Album{},
	}

	// Convert name to lowercase and replace spaces with underscores
	key := strings.ToLower(strings.ReplaceAll(r.FormValue("name"), " ", "_"))

	err := s.db.Write("bands", key, band)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated bands list
	s.handleBandsList(w, r)
} 


func (s *Server) handleBands(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodDelete:
        s.handleDeleteBand(w, r)
    case http.MethodPost:
        if strings.HasSuffix(r.URL.Path, "/albums") {
            s.handleAddAlbum(w, r)
        } else {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func (s *Server) handleDeleteBand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	encodedBandName := filepath.Base(r.URL.Path)
	decodedBandName, err := url.QueryUnescape(encodedBandName)
	if err != nil {
		http.Error(w, "Invalid band name encoding", http.StatusBadRequest)
		return
	}
	bandName := formatBandName(decodedBandName)
	
	err = s.db.Delete("bands", bandName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated bands list
	s.handleBandsList(w, r)
}

func (s *Server) handleAddAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// URL decode the band name first, then format it
	decodedBandName, err := url.QueryUnescape(parts[2])
	if err != nil {
		http.Error(w, "Invalid band name encoding", http.StatusBadRequest)
		return
	}
	bandName := formatBandName(decodedBandName)
	
	var band models.Band
	err = s.db.Read("bands", bandName, &band)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create new album
	album := models.Album{
		Name:  r.FormValue("albumName"),
		Year:  r.FormValue("year"),
		Genre: band.Genre,
	}

	// Add album to band
	band.Albums = append(band.Albums, album)

	// Update band in database
	err = s.db.Write("bands", bandName, band)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated bands list
	s.handleBandsList(w, r)
}

func (s *Server) handleBandsList(w http.ResponseWriter, r *http.Request) {
	var results []string
	var err error

	results, err = s.db.ReadAll("bands")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var bands []models.Band
	for _, record := range results {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(record), &data); err != nil {
			continue
		}
		band := models.Band{
			Name:    data["name"].(string),
			Country: data["country"].(string),
			Year:    fmt.Sprintf("%v", data["year"]),
			Genre:   data["genre"].(string),
		}
		
		if albumsData, ok := data["albums"].([]interface{}); ok {
			for _, albumData := range albumsData {
				if album, ok := albumData.(map[string]interface{}); ok {
					albumName := album["name"].(string)
					albumYear := fmt.Sprintf("%v", album["year"])
					albumGenre := album["genre"].(string)
					band.Albums = append(band.Albums, models.Album{
						Name:  albumName,
						Year:  albumYear,
						Genre: albumGenre,
					})
				}
			}
		}
		bands = append(bands, band)
	}

	// Sort bands by year in descending order
	sort.Slice(bands, func(i, j int) bool {
		yearI, _ := strconv.Atoi(bands[i].Year)
		yearJ, _ := strconv.Atoi(bands[j].Year)
		return yearI > yearJ
	})

	data := PageData{Bands: bands}
	s.templates.ExecuteTemplate(w, "bands", data)
}

func main() {
	// Create validators
	validators := map[string]db.ValidationFunc{
		"bands": func(data interface{}) error {
			switch v := data.(type) {
			case models.Band:
				if v.Name == "" {
					return fmt.Errorf("band name cannot be empty")
				}
				return nil
			default:
				return fmt.Errorf("invalid data type for band")
			}
		},
	}

	// Initialize the database
	database, err := db.New("data", &db.Options{Validators: validators})
	if err != nil {
		log.Fatal(err)
	}

	// Create and configure the server
	server := NewServer(database)

	// Define routes
	http.HandleFunc("/", server.handleIndex)
	http.HandleFunc("/addBand", server.handleAddBand)
	http.HandleFunc("/bands/", server.handleBands)

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
