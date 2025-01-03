package main

import (
	"log"
	"net/http"
	"os"

	"music-database/internal/handler"
	"music-database/internal/database"
)

func main() {
	// Set the project root directory explicitly
	projectRoot := "/home/ihor/Desktop/projects/music_database"
	
	// Change to the project root directory
	if err := os.Chdir(projectRoot); err != nil {
		log.Fatal(err)
	}

	// Print current working directory for debugging
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Current working directory: %s", cwd)
	
	// Check if templates directory exists
	if _, err := os.Stat("templates"); os.IsNotExist(err) {
		log.Fatal("templates directory not found")
	} else if err != nil {
		log.Fatal(err)
	}

	db, err := database.New("data")
	if err != nil {
		log.Fatal(err)
	}

	server := handler.NewServer(db)

	http.HandleFunc("/", server.HandleIndex)
	http.HandleFunc("/addBand", server.HandleAddBand)
	http.HandleFunc("/bands/", server.HandleDeleteBand) // This will handle both DELETE /bands/{name} and POST /bands/{name}/albums
	http.HandleFunc("/bands", server.HandleBands)
	http.HandleFunc("/add-album", server.HandleAddAlbum)
	http.HandleFunc("/bands-list", server.HandleBandsList)
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
