package health_test

import (
	"context"
	"fmt"
	"yp-go-short-url-service/internal/repository/mock"
	"yp-go-short-url-service/internal/service/health"

	"go.uber.org/mock/gomock"
)

// ExampleNewHealthCheckService демонстрирует создание нового сервиса проверки здоровья.
func ExampleNewHealthCheckService() {
	// Создаем контроллер для моков (в реальном приложении используйте реальный репозиторий)
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	// Создаем мок репозитория
	mockRepo := mock.NewMockURLRepositoryReader(ctrl)

	// Создаем сервис проверки здоровья
	service := health.NewHealthCheckService(mockRepo)

	// Сервис готов к использованию
	_ = service

	fmt.Println("HealthCheckService created successfully")
	// Output: HealthCheckService created successfully
}

// ExampleNewHealthCheckService_ping демонстрирует использование метода Ping для проверки соединения с базой данных.
func ExampleNewHealthCheckService_ping() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockRepo := mock.NewMockURLRepositoryReader(ctrl)
	service := health.NewHealthCheckService(mockRepo)

	ctx := context.Background()

	// Настраиваем ожидание успешного ответа
	mockRepo.EXPECT().Ping(ctx).Return(nil)

	// Выполняем проверку соединения
	err := service.Ping(ctx)
	if err != nil {
		fmt.Printf("Database connection check failed: %v\n", err)
		return
	}

	fmt.Println("Database connection is OK")
	// Output: Database connection is OK
}
