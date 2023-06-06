package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(dbURI string) *Storage {
	baseDir, _ := os.Getwd()
	filepaths := []string{filepath.Join(baseDir, "..", "db", "init.sql"), filepath.Join(baseDir, "cmd", "db", "init.sql")}

	var initScript string

	for _, filepath := range filepaths {
		if init, err := os.ReadFile(filepath); err != nil {
			// skip, continue searching
			fmt.Println("init script is not at: ", filepath)
		} else {
			initScript = string(init)
		}
	}

	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(initScript)
	if err != nil {
		panic(err)
	}

	return &Storage{db: db}
}

func (s *Storage) Close() error {
	return s.db.Close()
}
