package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	db "github.com/xsyro/goapi/internal/app/repo/sqlc"
)

//go:generate mockgen -source=repository.go -destination=../../mock/store.go -package=mocks
//go:generate mockgen -destination=../../mock/Tx.go -package=mocks github.com/jackc/pgx/v5 Tx
type Repository interface {
	PgxPoolInterface
	Queries
}

type Queries interface {
	db.Querier
	WithTx(tx pgx.Tx) *db.Queries
}

type PgxPoolInterface interface {
	Begin(context.Context) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Ping(context.Context) error
	Close()
}
