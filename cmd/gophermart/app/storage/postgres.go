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

func New() *Storage {
	dbUser := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	baseDir, _ := os.Getwd()
	initFilePath := filepath.Join(baseDir, "..", "db", "init.sql")

	init, err := os.ReadFile(initFilePath)
	if err != nil {
		panic(err)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
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
