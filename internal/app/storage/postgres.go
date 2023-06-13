package storage

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/vtcaregorodtcev/gophermarket/internal/helpers"
	"github.com/vtcaregorodtcev/gophermarket/internal/logger"
)

type Storage struct {
	db  *sql.DB
	log *zerolog.Logger
}

func New(dbURI string) *Storage {
	baseDir, _ := os.Getwd()
	log := logger.NewLogger("STORAGE")

	log.Info().Msgf("using postress db at: %s", dbURI)

	file := filepath.Join(baseDir, "..", "..", "migrations", "init.sql")

	isDevMode := *helpers.GetStringEnv("DEV_MODE", helpers.StringPtr("false")) == "true"

	if !isDevMode {
		file = filepath.Join(baseDir, "migrations", "init.sql")
	}

	init, err := os.ReadFile(file)
	if err != nil {
		log.Info().Msgf("init script is not found: %v", err)
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

	return &Storage{db: db, log: log}
}

func (s *Storage) Close() error {
	return s.db.Close()
}