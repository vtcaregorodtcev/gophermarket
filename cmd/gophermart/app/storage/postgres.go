package storage

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(dbURI string) *Storage {
	baseDir, _ := os.Getwd()
	initFilePath := filepath.Join(baseDir, "..", "db", "init.sql")

	init, err := os.ReadFile(initFilePath)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(string(init))
	if err != nil {
		panic(err)
	}

	return &Storage{db: db}
}

func (s *Storage) Close() error {
	return s.db.Close()
}
