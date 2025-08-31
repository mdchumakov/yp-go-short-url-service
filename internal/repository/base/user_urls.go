package base

import (
	"database/sql"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/repository/postgres"
	"yp-go-short-url-service/internal/repository/sqlite"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewUserURLsRepository(pool any) repository.UserURLsRepository {
	switch currentPool := pool.(type) {
	case *pgxpool.Pool:
		return postgres.NewUserURLsRepository(currentPool)
	case *sql.DB:
		return sqlite.NewUserURLsRepository(currentPool)
	default:
		panic("unsupported pool type")
	}
}
