package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUsersMockPool(t *testing.T) (pgxmock.PgxPoolIface, *usersRepository) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	repo := &usersRepository{pool: mock}
	return mock, repo
}

func TestUsersRepository_GetUsersCount_Success(t *testing.T) {
	mock, repo := setupUsersMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	expectedCount := int64(42)

	rows := pgxmock.NewRows([]string{"count"}).AddRow(expectedCount)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(rows)

	result, err := repo.GetUsersCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepository_GetUsersCount_ZeroCount(t *testing.T) {
	mock, repo := setupUsersMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	expectedCount := int64(0)

	rows := pgxmock.NewRows([]string{"count"}).AddRow(expectedCount)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(rows)

	result, err := repo.GetUsersCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepository_GetUsersCount_DatabaseError(t *testing.T) {
	mock, repo := setupUsersMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	expectedErr := errors.New("database connection error")

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnError(expectedErr)

	result, err := repo.GetUsersCount(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get users count")
	assert.Equal(t, int64(0), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepository_GetUsersCount_ScanError(t *testing.T) {
	mock, repo := setupUsersMockPool(t)
	defer mock.Close()

	ctx := context.Background()

	// Создаем строки с неправильным типом данных для вызова ошибки сканирования
	rows := pgxmock.NewRows([]string{"count"}).
		AddRow("invalid_count") // строка вместо int64

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(rows)

	result, err := repo.GetUsersCount(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scan users count")
	assert.Equal(t, int64(0), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepository_GetUsersCount_NoRowsReturned(t *testing.T) {
	mock, repo := setupUsersMockPool(t)
	defer mock.Close()

	ctx := context.Background()

	// Создаем пустые строки (нет строк в результате)
	rows := pgxmock.NewRows([]string{"count"})

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(rows)

	result, err := repo.GetUsersCount(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no rows returned for users count")
	assert.Equal(t, int64(0), result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersRepository_GetUsersCount_LargeCount(t *testing.T) {
	mock, repo := setupUsersMockPool(t)
	defer mock.Close()

	ctx := context.Background()
	expectedCount := int64(1000000)

	rows := pgxmock.NewRows([]string{"count"}).AddRow(expectedCount)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(rows)

	result, err := repo.GetUsersCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
