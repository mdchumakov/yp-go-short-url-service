package health

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"yp-go-short-url-service/internal/repository/mock"
)

func TestNewHealthCheckService(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepositoryReader(ctrl)

	// Создаем сервис через тестовый конструктор
	service := NewHealthCheckService(mockRepo)

	// Проверяем, что сервис создан корректно
	assert.NotNil(t, service)
	assert.IsType(t, &healthCheckService{}, service)
}

func Test_healthCheckService_Ping(t *testing.T) {
	// Создаем контроллер для моков
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepositoryReader(ctrl)

	// Создаем сервис
	service := &healthCheckService{
		repository: mockRepo,
	}

	ctx := context.Background()

	t.Run("successful ping", func(t *testing.T) {
		// Ожидаем успешный вызов Ping
		mockRepo.EXPECT().
			Ping(ctx).
			Return(nil)

		// Вызываем метод
		err := service.Ping(ctx)

		// Проверяем результат
		assert.NoError(t, err)
	})

	t.Run("ping with error", func(t *testing.T) {
		expectedErr := assert.AnError

		// Ожидаем вызов Ping с ошибкой
		mockRepo.EXPECT().
			Ping(ctx).
			Return(expectedErr)

		// Вызываем метод
		err := service.Ping(ctx)

		// Проверяем результат
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
