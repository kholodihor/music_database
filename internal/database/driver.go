package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"music-database/pkg/models"
)

type Driver struct {
	Dir string
}

type Query struct {
	Field    string
	Operator string
	Value    string
}

func New(dir string) (*Driver, error) {
	return &Driver{Dir: dir}, nil
}

func (d *Driver) Query(collection string, query Query) ([]models.Band, error) {
	var bands []models.Band
	files, err := ioutil.ReadDir(filepath.Join(d.Dir, collection))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join(d.Dir, collection, file.Name()))
		if err != nil {
			return nil, err
		}

		var band models.Band
		if err := json.Unmarshal(data, &band); err != nil {
			return nil, err
		}

		if query.Field == "" || (query.Field == "genre" && strings.EqualFold(band.Genre, query.Value)) {
			bands = append(bands, band)
		}
	}

	return bands, nil
}

func (d *Driver) Save(collection string, id string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Join(d.Dir, collection)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(dir, fmt.Sprintf("%s.json", id)), jsonData, 0644)
}

func (d *Driver) Delete(collection string, id string) error {
	return os.Remove(filepath.Join(d.Dir, collection, fmt.Sprintf("%s.json", id)))
}

func (d *Driver) Get(collection string, id string) (models.Band, error) {
	var band models.Band
	data, err := ioutil.ReadFile(filepath.Join(d.Dir, collection, fmt.Sprintf("%s.json", id)))
	if err != nil {
		return band, err
	}

	if err := json.Unmarshal(data, &band); err != nil {
		return band, err
	}

	return band, nil
}
