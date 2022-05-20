package databasesupport

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
)

func Open(url string) (*sql.DB, error) {
	return sql.Open("postgres", url)
}

func WithTransaction(db *sql.DB, ctx context.Context, options sql.TxOptions, f func(transaction *sql.Tx) (interface{}, error)) (interface{}, error) {
	tx, beginErr := db.BeginTx(ctx, &options)
	if beginErr != nil {
		return nil, beginErr
	}
	response, err := f(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return nil, rbErr
		}
		return nil, err
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return nil, commitErr
	}
	return response, nil
}
