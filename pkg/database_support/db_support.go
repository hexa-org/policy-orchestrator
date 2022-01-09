package database_support

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func Open(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		panic(err)
	}
	return db, err
}
