package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"music-database/internal/database"
	"music-database/pkg/models"
)

type Server struct {
	db        *database.Driver
	templates *template.Template
}

type PageData struct {
	Bands []models.Band
}

func formatBandName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "_"))
}

func NewServer(database *database.Driver) *Server {
	funcMap := template.FuncMap{
		"formatBandName": formatBandName,
	}
	
	// Get absolute path to templates
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	templatePath := filepath.Join(cwd, "templates", "*.html")
	
	templates := template.New("").Funcs(funcMap)
	templates = template.Must(templates.ParseGlob(templatePath))
	return &Server{
		db:        database,
		templates: templates,
	}
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	results, err := s.db.Query("bands", database.Query{Field: "genre", Operator: "eq", Value: "Progressive Rock"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Bands: results,
	}

	s.templates.ExecuteTemplate(w, "layout.html", data)
}

func (s *Server) HandleAddBand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	band := models.Band{
		Name:    r.FormValue("name"),
		Country: r.FormValue("country"),
		Genre:   r.FormValue("genre"),
	}

	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}
	band.Year = year

	if err := s.db.Save("bands", formatBandName(band.Name), band); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Refresh the bands list
	results, err := s.db.Query("bands", database.Query{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Bands: results,
	}

	s.templates.ExecuteTemplate(w, "bands", data)
}

func (s *Server) HandleBands(w http.ResponseWriter, r *http.Request) {
	results, err := s.db.Query("bands", database.Query{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}

func (s *Server) HandleDeleteBand(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	bandName := parts[2]
	
	if r.Method == http.MethodDelete {
		// Handle band deletion
		if err := s.db.Delete("bands", formatBandName(bandName)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Refresh the bands list
		results, err := s.db.Query("bands", database.Query{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := PageData{
			Bands: results,
		}

		s.templates.ExecuteTemplate(w, "bands", data)
	} else if r.Method == http.MethodPost && len(parts) == 4 && parts[3] == "albums" {
		// Handle album addition
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		year, err := strconv.Atoi(r.FormValue("year"))
		if err != nil {
			http.Error(w, "Invalid year", http.StatusBadRequest)
			return
		}

		album := models.Album{
			Name:  r.FormValue("albumName"),
			Year:  year,
			Genre: r.FormValue("genre"),
		}

		// Get the band
		band, err := s.db.Get("bands", bandName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Add the album
		band.Albums = append(band.Albums, album)
		sort.Slice(band.Albums, func(i, j int) bool {
			return band.Albums[i].Year < band.Albums[j].Year
		})

		// Save the updated band
		if err := s.db.Save("bands", bandName, band); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Refresh the bands list
		results, err := s.db.Query("bands", database.Query{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := PageData{
			Bands: results,
		}

		s.templates.ExecuteTemplate(w, "bands", data)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) HandleAddAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	bandName := parts[2]

	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	album := models.Album{
		Name:  r.FormValue("albumName"),
		Year:  year,
		Genre: r.FormValue("genre"),
	}

	// Get the band
	band, err := s.db.Get("bands", bandName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add the album
	band.Albums = append(band.Albums, album)
	sort.Slice(band.Albums, func(i, j int) bool {
		return band.Albums[i].Year < band.Albums[j].Year
	})

	// Save the updated band
	if err := s.db.Save("bands", bandName, band); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Refresh the bands list
	results, err := s.db.Query("bands", database.Query{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Bands: results,
	}

	s.templates.ExecuteTemplate(w, "bands", data)
}

func (s *Server) HandleBandsList(w http.ResponseWriter, r *http.Request) {
	results, err := s.db.Query("bands", database.Query{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Bands: results,
	}

	s.templates.ExecuteTemplate(w, "bands-list.html", data)
}
