package seed

import (
	"fmt"

	"music-database/db"
	"music-database/models"
)

// Database initializes the database with sample prog rock bands
func Database(database *db.Driver) error {
	// Example usage - Pink Floyd
	pinkFloyd := models.Band{
		Name:    "Pink Floyd",
		Country: "United Kingdom",
		Year:    "1965",
		Genre:   "Progressive Rock",
		Albums: []models.Album{
			{Name: "The Dark Side of the Moon", Year: "1973", Genre: "Progressive Rock"},
			{Name: "Wish You Were Here", Year: "1975", Genre: "Progressive Rock"},
			{Name: "Animals", Year: "1977", Genre: "Progressive Rock"},
			{Name: "The Wall", Year: "1979", Genre: "Progressive Rock"},
		},
	}

	// Write Pink Floyd
	if err := database.Write("bands", "pink_floyd", pinkFloyd); err != nil {
		return fmt.Errorf("error writing Pink Floyd: %w", err)
	}

	// Add more albums to Pink Floyd
	updates := map[string]interface{}{
		"albums": []models.Album{
			{Name: "The Dark Side of the Moon", Year: "1973", Genre: "Progressive Rock"},
			{Name: "Wish You Were Here", Year: "1975", Genre: "Progressive Rock"},
			{Name: "Animals", Year: "1977", Genre: "Progressive Rock"},
			{Name: "The Wall", Year: "1979", Genre: "Progressive Rock"},
			{Name: "Meddle", Year: "1971", Genre: "Progressive Rock"},
			{Name: "Atom Heart Mother", Year: "1970", Genre: "Progressive Rock"},
		},
	}

	if err := database.Update("bands", "pink_floyd", updates); err != nil {
		return fmt.Errorf("error updating Pink Floyd: %w", err)
	}

	// Batch write other prog rock bands
	bands := map[string]interface{}{
		"king_crimson": models.Band{
			Name:    "King Crimson",
			Country: "United Kingdom",
			Year:    "1968",
			Genre:   "Progressive Rock",
			Albums: []models.Album{
				{Name: "In the Court of the Crimson King", Year: "1969", Genre: "Progressive Rock"},
				{Name: "Red", Year: "1974", Genre: "Progressive Rock"},
				{Name: "Discipline", Year: "1981", Genre: "Progressive Rock"},
				{Name: "Larks' Tongues in Aspic", Year: "1973", Genre: "Progressive Rock"},
			},
		},
		"yes": models.Band{
			Name:    "Yes",
			Country: "United Kingdom",
			Year:    "1968",
			Genre:   "Progressive Rock",
			Albums: []models.Album{
				{Name: "Close to the Edge", Year: "1972", Genre: "Progressive Rock"},
				{Name: "Fragile", Year: "1971", Genre: "Progressive Rock"},
				{Name: "The Yes Album", Year: "1971", Genre: "Progressive Rock"},
				{Name: "Relayer", Year: "1974", Genre: "Progressive Rock"},
			},
		},
	}

	if err := database.BatchWrite("bands", bands); err != nil {
		return fmt.Errorf("error batch writing: %w", err)
	}

	return nil
}
