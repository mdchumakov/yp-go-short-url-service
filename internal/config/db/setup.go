package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type SetupParams struct {
	PostgresDSN      string
	PGMigrationsPath string
	SQLiteDSN        string
}

func (s *SetupParams) Validate() error {
	if s.SQLiteDSN == "" && s.PostgresDSN == "" {
		return errors.New("db connection string is empty")
	} else if s.PostgresDSN == "" && s.PGMigrationsPath == "" {
		return errors.New("migrations path is empty")
	}

	return nil
}

func Setup(
	ctx context.Context,
	logger *zap.SugaredLogger,
	params *SetupParams,
) any {

	if err := params.Validate(); err != nil {
		return err
	}

	pgPool, err := setupPostgres(ctx, logger, params)
	if err == nil {
		logger.Info("PostgreSQL успешно инициализирован")
		return pgPool
	}
	logger.Warnw("PostgreSQL недоступен, переключаемся на SQLite", "error", err)

	sqlitePool, err := setupSQLite(logger, params)
	if err == nil {
		return sqlitePool
	}

	logger.Fatalw("не удалось инициализировать SQLite", "error", err)
	return err
}

func setupPostgres(
	ctx context.Context,
	logger *zap.SugaredLogger,
	params *SetupParams,
) (*pgxpool.Pool, error) {
	pgPool, err := InitPostgresDB(
		ctx,
		params.PostgresDSN,
	)
	if err != nil {
		return nil, err
	}

	if err = RunMigrations(logger, pgPool, params.PGMigrationsPath); err != nil {
		return nil, err
	}
	return pgPool, nil
}

func setupSQLite(
	logger *zap.SugaredLogger,
	params *SetupParams,
) (*sql.DB, error) {
	sqliteDB, err := SetupSQLiteDB(params.SQLiteDSN, logger)
	if err != nil {
		return nil, err
	}

	return sqliteDB, nil
}
