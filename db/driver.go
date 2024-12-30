package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jcelliott/lumber"
)

const Version = "1.0.1"

type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}

	// ValidationFunc is a function type for data validation
	ValidationFunc func(interface{}) error

	Driver struct {
		mutex      sync.Mutex
		mutexes    map[string]*sync.Mutex
		dir        string
		log        Logger
		validators map[string]ValidationFunc
		stats      *CollectionStats
	}

	// CollectionStats tracks database statistics
	CollectionStats struct {
		mutex       sync.Mutex
		Operations  map[string]int    // Count of operations by type
		AccessTime  map[string]int64  // Last access time by collection
		RecordCount map[string]int    // Number of records by collection
	}
)

type Options struct {
	Logger
	Validators map[string]ValidationFunc
}

// Query represents a simple query structure
type Query struct {
	Field    string
	Operator string
	Value    interface{}
}

func New(dir string, options *Options) (*Driver, error) {
	dir = filepath.Clean(dir)

	opts := Options{}
	if options != nil {
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
	}

	driver := &Driver{
		dir:        dir,
		mutexes:    make(map[string]*sync.Mutex),
		log:        opts.Logger,
		validators: opts.Validators,
		stats: &CollectionStats{
			Operations:  make(map[string]int),
			AccessTime:  make(map[string]int64),
			RecordCount: make(map[string]int),
		},
	}

	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using existing database at '%s'\n", dir)
		return driver, nil
	}

	opts.Logger.Debug("Creating database at '%s'\n", dir)
	return driver, os.MkdirAll(dir, 0755)
}

// AddValidator adds a validation function for a specific collection
func (d *Driver) AddValidator(collection string, validator ValidationFunc) {
	if d.validators == nil {
		d.validators = make(map[string]ValidationFunc)
	}
	d.validators[collection] = validator
}

func (d *Driver) Write(collection, resource string, data interface{}) error {
	if collection == "" {
		return &DbError{Code: ErrCodeInvalidInput, Message: "collection cannot be empty"}
	}
	if resource == "" {
		return &DbError{Code: ErrCodeInvalidInput, Message: "resource cannot be empty"}
	}

	// Run validator if exists
	if validator, exists := d.validators[collection]; exists {
		if err := validator(data); err != nil {
			return &DbError{Code: ErrCodeInvalidInput, Message: "validation failed", Err: err}
		}
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	finalPath := filepath.Join(dir, resource+".json")
	tmpPath := finalPath + ".tmp"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return &DbError{Code: ErrCodeInternal, Message: "failed to create directory", Err: err}
	}

	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return &DbError{Code: ErrCodeInternal, Message: "failed to marshal data", Err: err}
	}

	b = append(b, byte('\n'))
	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return &DbError{Code: ErrCodeInternal, Message: "failed to write file", Err: err}
	}

	err = os.Rename(tmpPath, finalPath)
	if err != nil {
		return &DbError{Code: ErrCodeInternal, Message: "failed to rename file", Err: err}
	}

	// Update stats
	d.updateStats(collection, "write")
	return nil
}

// Update updates an existing resource in the collection
func (d *Driver) Update(collection, resource string, updates map[string]interface{}) error {
	if collection == "" {
		return &DbError{Code: ErrCodeInvalidInput, Message: "collection cannot be empty"}
	}
	if resource == "" {
		return &DbError{Code: ErrCodeInvalidInput, Message: "resource cannot be empty"}
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	filePath := filepath.Join(d.dir, collection, resource+".json")
	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &DbError{Code: ErrCodeNotFound, Message: fmt.Sprintf("resource '%s' not found in collection '%s'", resource, collection)}
		}
		return &DbError{Code: ErrCodeInternal, Message: "failed to read file", Err: err}
	}

	var data map[string]interface{}
	if err := json.Unmarshal(file, &data); err != nil {
		return &DbError{Code: ErrCodeInternal, Message: "failed to unmarshal data", Err: err}
	}

	// Apply updates
	for key, value := range updates {
		data[key] = value
	}

	// Validate updated data
	if validator, exists := d.validators[collection]; exists {
		if err := validator(data); err != nil {
			return &DbError{Code: ErrCodeInvalidInput, Message: "validation failed", Err: err}
		}
	}

	// Write updated data
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return &DbError{Code: ErrCodeInternal, Message: "failed to marshal data", Err: err}
	}

	b = append(b, byte('\n'))
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return &DbError{Code: ErrCodeInternal, Message: "failed to write file", Err: err}
	}

	if err := os.Rename(tmpPath, filePath); err != nil {
		return &DbError{Code: ErrCodeInternal, Message: "failed to rename file", Err: err}
	}

	// Update stats
	d.updateStats(collection, "update")
	return nil
}

// BatchWrite performs multiple write operations in a single transaction
func (d *Driver) BatchWrite(collection string, items map[string]interface{}) error {
	if collection == "" {
		return &DbError{Code: ErrCodeInvalidInput, Message: "collection cannot be empty"}
	}

	for resource, data := range items {
		if err := d.Write(collection, resource, data); err != nil {
			return &DbError{Code: ErrCodeInternal, Message: "batch write failed", Err: err}
		}
	}

	return nil
}

func (d *Driver) Read(collection, resource string, data interface{}) error {
	if collection == "" {
		return &DbError{Code: ErrCodeInvalidInput, Message: "collection cannot be empty"}
	}
	if resource == "" {
		return &DbError{Code: ErrCodeInvalidInput, Message: "resource cannot be empty"}
	}

	filePath := filepath.Join(d.dir, collection, resource+".json")
	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &DbError{Code: ErrCodeNotFound, Message: fmt.Sprintf("resource '%s' not found in collection '%s'", resource, collection)}
		}
		return &DbError{Code: ErrCodeInternal, Message: "failed to read file", Err: err}
	}

	if err := json.Unmarshal(file, data); err != nil {
		return &DbError{Code: ErrCodeInternal, Message: "failed to unmarshal data", Err: err}
	}

	// Update stats
	d.updateStats(collection, "read")
	return nil
}

// Delete removes a resource from the collection
func (d *Driver) Delete(collection, resource string) error {
	if collection == "" {
		return &DbError{Code: ErrCodeInvalidInput, Message: "collection cannot be empty"}
	}
	if resource == "" {
		return &DbError{Code: ErrCodeInvalidInput, Message: "resource cannot be empty"}
	}

	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	filePath := filepath.Join(d.dir, collection, resource+".json")
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return &DbError{Code: ErrCodeNotFound, Message: fmt.Sprintf("resource '%s' not found in collection '%s'", resource, collection)}
		}
		return &DbError{Code: ErrCodeInternal, Message: "failed to delete file", Err: err}
	}

	// Update stats
	d.updateStats(collection, "delete")
	return nil
}

// Query performs a simple query operation on a collection
func (d *Driver) Query(collection string, query Query) ([]interface{}, error) {
	records, err := d.ReadAll(collection)
	if err != nil {
		return nil, err
	}

	var results []interface{}
	for _, record := range records {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(record), &data); err != nil {
			continue
		}

		if value, exists := data[query.Field]; exists {
			matches := false
			switch query.Operator {
			case "eq":
				matches = reflect.DeepEqual(value, query.Value)
			case "gt":
				matches = compareValues(value, query.Value) > 0
			case "lt":
				matches = compareValues(value, query.Value) < 0
			}
			if matches {
				results = append(results, data)
			}
		}
	}

	return results, nil
}

func (d *Driver) ReadAll(collection string) ([]string, error) {
	dir := filepath.Join(d.dir, collection)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var records []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			records = append(records, string(data))
		}
	}

	return records, nil
}

// GetStats returns the current statistics for a collection
func (d *Driver) GetStats(collection string) map[string]interface{} {
	d.stats.mutex.Lock()
	defer d.stats.mutex.Unlock()

	return map[string]interface{}{
		"operations":   d.stats.Operations[collection],
		"access_time":  time.Unix(d.stats.AccessTime[collection], 0),
		"record_count": d.stats.RecordCount[collection],
	}
}

func (d *Driver) updateStats(collection, operation string) {
	d.stats.mutex.Lock()
	defer d.stats.mutex.Unlock()

	d.stats.Operations[collection]++
	d.stats.AccessTime[collection] = time.Now().Unix()
	
	// Update record count
	files, _ := os.ReadDir(filepath.Join(d.dir, collection))
	d.stats.RecordCount[collection] = len(files)
}

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	m, ok := d.mutexes[collection]
	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}
	return m
}

// Helper function to compare values
func compareValues(a, b interface{}) int {
	switch v1 := a.(type) {
	case float64:
		if v2, ok := b.(float64); ok {
			if v1 < v2 {
				return -1
			} else if v1 > v2 {
				return 1
			}
			return 0
		}
	case string:
		if v2, ok := b.(string); ok {
			return strings.Compare(v1, v2)
		}
	}
	return 0
}
