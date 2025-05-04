package repo

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/xsyro/goapi/internal/app/repo/sqlc"
	"github.com/xsyro/goapi/internal/utils"
)

type Store struct {
	*pgxpool.Pool
	*db.Queries
}

var (
	repo   *Store
	dbOnce sync.Once
)

func NewRepo(dataSourceName string) (Repository, error) {
	var initErr error

	dbOnce.Do(func() {
		ctx := context.Background()
		dbpool, err := pgxpool.New(ctx, dataSourceName)
		if err != nil {
			initErr = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}

		if err := dbpool.Ping(ctx); err != nil {
			initErr = fmt.Errorf("error connecting to database: %w", err)
			dbpool.Close() // If there's an error, close the pool before exiting
			return
		}

		repo = &Store{
			Pool:    dbpool,
			Queries: db.New(dbpool),
		}
	})

	return repo, initErr
}

func WrapDbError(err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return utils.WrapDbError(err, utils.ErrNotFound)
	case checkPgError(err, pgerrcode.UniqueViolation):
		return utils.WrapDbError(err, utils.ErrDuplicate)
	default:
		return utils.WrapDbError(err, utils.ErrInternal)
	}
}

func checkPgError(err error, code string) bool {
	var pqErr *pgconn.PgError
	if errors.As(err, &pqErr) {
		return pqErr.Code == code
	}
	return false
}
