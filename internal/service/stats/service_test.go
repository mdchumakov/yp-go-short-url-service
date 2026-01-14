package stats

import (
	"context"
	"errors"
	"testing"

	"yp-go-short-url-service/internal/repository/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки репозиториев
	mockUserRepo := mock.NewMockUserRepositoryReader(ctrl)
	mockURLsRepo := mock.NewMockURLRepositoryReader(ctrl)

	// Создаем сервис через тестовый конструктор
	service := New(mockUserRepo, mockURLsRepo)

	// Проверяем, что сервис создан корректно
	assert.NotNil(t, service)
	assert.IsType(t, &serviceImpl{}, service)
}

func Test_serviceImpl_GetTotalURLsCount(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки репозиториев
	mockUserRepo := mock.NewMockUserRepositoryReader(ctrl)
	mockURLsRepo := mock.NewMockURLRepositoryReader(ctrl)

	// Создаем сервис
	service := &serviceImpl{
		userRepo: mockUserRepo,
		urlsRepo: mockURLsRepo,
	}

	ctx := context.Background()

	t.Run("successful get total URLs count", func(t *testing.T) {
		expectedCount := int64(100)

		// Ожидаем успешный вызов GetTotalCount
		mockURLsRepo.EXPECT().
			GetTotalCount(ctx).
			Return(expectedCount, nil)

		// Вызываем метод
		count, err := service.GetTotalURLsCount(ctx)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, count)
	})

	t.Run("get total URLs count with error", func(t *testing.T) {
		expectedErr := errors.New("database error")

		// Ожидаем вызов GetTotalCount с ошибкой
		mockURLsRepo.EXPECT().
			GetTotalCount(ctx).
			Return(int64(0), expectedErr)

		// Вызываем метод
		count, err := service.GetTotalURLsCount(ctx)

		// Проверяем результат
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("get total URLs count with zero count", func(t *testing.T) {
		expectedCount := int64(0)

		// Ожидаем успешный вызов GetTotalCount с нулевым результатом
		mockURLsRepo.EXPECT().
			GetTotalCount(ctx).
			Return(expectedCount, nil)

		// Вызываем метод
		count, err := service.GetTotalURLsCount(ctx)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, count)
	})
}

func Test_serviceImpl_GetTotalUsersCount(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки репозиториев
	mockUserRepo := mock.NewMockUserRepositoryReader(ctrl)
	mockURLsRepo := mock.NewMockURLRepositoryReader(ctrl)

	// Создаем сервис
	service := &serviceImpl{
		userRepo: mockUserRepo,
		urlsRepo: mockURLsRepo,
	}

	ctx := context.Background()

	t.Run("successful get total users count", func(t *testing.T) {
		expectedCount := int64(50)

		// Ожидаем успешный вызов GetUsersCount
		mockUserRepo.EXPECT().
			GetUsersCount(ctx).
			Return(expectedCount, nil)

		// Вызываем метод
		count, err := service.GetTotalUsersCount(ctx)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, count)
	})

	t.Run("get total users count with error", func(t *testing.T) {
		expectedErr := errors.New("database connection failed")

		// Ожидаем вызов GetUsersCount с ошибкой
		mockUserRepo.EXPECT().
			GetUsersCount(ctx).
			Return(int64(0), expectedErr)

		// Вызываем метод
		count, err := service.GetTotalUsersCount(ctx)

		// Проверяем результат
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("get total users count with zero count", func(t *testing.T) {
		expectedCount := int64(0)

		// Ожидаем успешный вызов GetUsersCount с нулевым результатом
		mockUserRepo.EXPECT().
			GetUsersCount(ctx).
			Return(expectedCount, nil)

		// Вызываем метод
		count, err := service.GetTotalUsersCount(ctx)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, count)
	})
}
