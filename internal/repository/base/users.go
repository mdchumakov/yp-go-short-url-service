package base

import (
	"database/sql"

	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/repository/postgres"
	"yp-go-short-url-service/internal/repository/sqlite"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewUsersRepository создает новый репозиторий для работы с пользователями в зависимости от типа пула соединений.
// Поддерживает PostgreSQL (pgxpool.Pool) и SQLite (*sql.DB). Возвращает соответствующую реализацию интерфейса UserRepository.
func NewUsersRepository(pool any) repository.UserRepository {
	switch currentPool := pool.(type) {
	case *pgxpool.Pool:
		return postgres.NewUsersRepository(currentPool)
	case *sql.DB:
		return sqlite.NewUsersRepository(currentPool)
	default:
		panic("unsupported pool type")
	}
}
