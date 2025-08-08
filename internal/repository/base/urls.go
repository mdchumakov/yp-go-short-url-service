package base

import (
	"database/sql"

	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/repository/postgres"
	"yp-go-short-url-service/internal/repository/sqlite"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewURLsRepository(pool any) repository.URLRepository {
	switch currentPool := pool.(type) {
	case *pgxpool.Pool:
		return postgres.NewURLsRepository(currentPool)
	case *sql.DB:
		return sqlite.NewURLsRepository(currentPool)
	default:
		panic("unsupported pool type")
	}
}
