# ColdDB - Simple JSON Document Database

ColdDB is a lightweight, file-based JSON document database written in Go. It provides basic CRUD operations, querying capabilities, and collection statistics.

## Features

- File-based JSON storage
- CRUD operations (Create, Read, Update, Delete)
- Batch write operations
- Query support with basic operators (eq, gt, lt)
- Data validation hooks
- Collection statistics
- Thread-safe operations
- Custom error types

## Installation

```bash
go get github.com/kholodihor/colddb
```

## Usage

```go
import (
    "github.com/kholodihor/colddb/db"
    "github.com/kholodihor/colddb/models"
)

// Initialize database
database, err := db.New("./data", nil)
if err != nil {
    log.Fatal(err)
}

// Create a record
band := models.Band{
    Name:    "Metallica",
    Country: "USA",
    Year:    "1981",
    Genre:   "Metal",
    Albums: []models.Album{
        {Name: "Master of Puppets", Year: 1986},
    },
}

// Write to database
err = database.Write("bands", "metallica", band)

// Update a record
updates := map[string]interface{}{
    "genre": "Thrash Metal",
}
err = database.Update("bands", "metallica", updates)

// Query records
results, err := database.Query("bands", db.Query{
    Field:    "genre",
    Operator: "eq",
    Value:    "Thrash Metal",
})

// Get collection statistics
stats := database.GetStats("bands")
```

## License

MIT License
