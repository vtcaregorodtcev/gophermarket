package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/vtcaregorodtcev/gophermarket/cmd/gophermart/pkg/logger"
)

type Storage struct {
	db  *sql.DB
	log *zerolog.Logger
}

func New(dbURI string) *Storage {
	baseDir, _ := os.Getwd()
	filepaths := []string{
		// works from make script
		filepath.Join(baseDir, "..", "db", "init.sql"),
		// works from binary
		filepath.Join(baseDir, "cmd", "db", "init.sql"),
	}

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

	return &Storage{db: db, log: logger.NewLogger("STORAGE")}
}

func (s *Storage) Close() error {
	return s.db.Close()
}
