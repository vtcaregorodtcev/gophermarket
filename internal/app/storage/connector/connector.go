package connector

import "database/sql"

type Connector interface {
	Connect(dbURI string) (*sql.DB, error)
}
