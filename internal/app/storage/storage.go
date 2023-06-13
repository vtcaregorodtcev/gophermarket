package storage

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/vtcaregorodtcev/gophermarket/internal/app/storage/connector"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) init() error {
	baseDir, _ := os.Getwd()

	log.Printf("using PostgreSQL DB")

	file := filepath.Join(baseDir, "..", "..", "migrations", "init.sql")

	isDevMode := os.Getenv("DEV_MODE") == "true"

	if !isDevMode {
		file = filepath.Join(baseDir, "migrations", "init.sql")
	}

	init, err := os.ReadFile(file)
	if err != nil {
		log.Printf("init script is not found: %v", err)
		return err
	}

	err = s.db.Ping()
	if err != nil {
		return err
	}

	_, err = s.db.Exec(string(init))
	return err
}

func New(dbURI string, connector connector.Connector) (*Storage, error) {
	db, err := connector.Connect(dbURI)
	if err != nil {
		return nil, err
	}

	storage := &Storage{db: db}
	err = storage.init()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
