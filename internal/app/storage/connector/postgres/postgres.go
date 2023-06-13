package postgres

import "database/sql"

type PGConnector struct{}

func (p *PGConnector) Connect(dbURI string) (*sql.DB, error) {
	return sql.Open("postgres", dbURI)
}
