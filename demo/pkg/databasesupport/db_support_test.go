package databasesupport

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	db, _ := Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	err := db.Ping()
	assert.NoError(t, err)
}

type Record struct {
	ID   string
	Name string
}

func TestWithTransaction(t *testing.T) {
	db, _ := Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	ctx := context.Background()
	response, err := WithTransaction(db, ctx, sql.TxOptions{}, func(*sql.Tx) (interface{}, error) {
		var id string
		first := db.QueryRow("insert into integrations (name, provider, key) values ($1, $2, $3) returning id",
			"aName", "aProvider", []byte("aKey")).Scan(&id)
		if first != nil {
			return nil, first
		}
		if id == "" {
			return nil, errors.New("unable to create integration record")
		}

		var record Record
		row := db.QueryRow("select id, name from integrations where id=$1 for update", id)
		second := row.Scan(&record.ID, &record.Name)
		if second != nil {
			return nil, second
		}
		if record.ID == "" {
			return nil, errors.New("unable to find integration record")
		}
		var appId string
		third := db.QueryRow(`insert into applications (integration_id, object_id, name, description) values ($1, $2, $3, $4) returning id`,
			record.ID, "anObjectId", "aName", "aDescription").Scan(&appId)
		if third != nil {
			return nil, third
		}
		return appId, nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, response.(string))
	_ = db.Close()
}

func TestWithTransaction_RollsBack(t *testing.T) {
	db, _ := Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	ctx := context.Background()
	_, err := WithTransaction(db, ctx, sql.TxOptions{}, func(*sql.Tx) (interface{}, error) {
		var id string
		first := db.QueryRow("insert into integrations (name, provider, key) values ($1, $2, $3) returning id",
			"aName", "aProvider", []byte("aKey")).Scan(&id)
		if first != nil {
			return nil, first
		}

		var appId string
		third := db.QueryRow(`insert into applications (id, integration_id, object_id, name, description) values ($1, $2, $3, $4, $5) returning id`,
			"aBadId", id, "anObjectId", "aName", "aDescription").Scan(&id)
		if third != nil {
			return nil, third
		}
		return appId, nil
	})
	assert.Contains(t, err.Error(), "invalid input syntax for type uuid:")
	row := db.QueryRow("select count from integrations")
	var count int
	_ = row.Scan(&count)
	assert.Equal(t, 0, count)
	_ = db.Close()
}

func TestWithTransaction_BeginTransactionError(t *testing.T) {
	db, _ := Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	_ = db.Close()
	ctx := context.Background()
	_, err := WithTransaction(db, ctx, sql.TxOptions{}, func(tx *sql.Tx) (interface{}, error) {
		return nil, nil
	})
	assert.Equal(t, err.Error(), "sql: database is closed")
}

func TestWithTransaction_RollBackError(t *testing.T) {
	db, _ := Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	ctx := context.Background()
	_, err := WithTransaction(db, ctx, sql.TxOptions{}, func(tx *sql.Tx) (interface{}, error) {
		_ = tx.Rollback()
		return nil, errors.New("oops")
	})
	assert.Equal(t, err.Error(), "sql: transaction has already been committed or rolled back")
}

func TestWithTransaction_CommitTransactionError(t *testing.T) {
	db, _ := Open("postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable")
	ctx := context.Background()
	_, err := WithTransaction(db, ctx, sql.TxOptions{}, func(tx *sql.Tx) (interface{}, error) {
		_ = tx.Commit()
		return nil, nil
	})
	assert.Equal(t, err.Error(), "sql: transaction has already been committed or rolled back")
}
