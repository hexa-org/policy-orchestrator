package database_support

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpen(t *testing.T) {
	db, err := Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	err = db.Ping()
	assert.NoError(t, err)
}
