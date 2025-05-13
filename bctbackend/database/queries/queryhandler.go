package queries

import (
	"database/sql"
)

type QueryHandler interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
}
