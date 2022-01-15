package databasesupport

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func Open(url string) (*sql.DB, error) {
	return sql.Open("postgres", url)
}
