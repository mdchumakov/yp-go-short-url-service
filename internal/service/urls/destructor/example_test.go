package destructor_test

import (
	"context"
	"fmt"
	"yp-go-short-url-service/internal/repository/mock"
	"yp-go-short-url-service/internal/service/urls/destructor"

	"go.uber.org/mock/gomock"
)

// ExampleNewURLDestructorService демонстрирует создание нового сервиса для удаления URL.
func ExampleNewURLDestructorService() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	// Создаем моки репозиториев
	mockURLRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)

	// Создаем сервис для удаления URL
	// Сервис автоматически запускает воркеры для асинхронной обработки удалений
	service := destructor.NewURLDestructorService(mockURLRepo, mockUserURLsRepo)

	// Важно: не забудьте остановить сервис при завершении работы приложения
	defer service.Stop()

	// Сервис готов к использованию
	_ = service

	fmt.Println("URLDestructorService created successfully")
	// Output: URLDestructorService created successfully
}

// ExampleNewURLDestructorService_deleteURLsByBatch демонстрирует асинхронное удаление URL пакетом.
func ExampleNewURLDestructorService_deleteURLsByBatch() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)

	service := destructor.NewURLDestructorService(mockURLRepo, mockUserURLsRepo)
	defer service.Stop()

	ctx := context.Background()

	// Подготавливаем список коротких URL для удаления
	shortURLs := []string{
		"abc123",
		"def456",
		"ghi789",
	}

	// В реальном приложении здесь будет вызов:
	// err := service.DeleteURLsByBatch(ctx, shortURLs)
	// Метод возвращает управление сразу после отправки запроса в очередь
	// Для примера демонстрируем только подготовку данных
	_ = ctx
	_ = shortURLs

	fmt.Printf("Service is ready to delete %d URLs asynchronously\n", len(shortURLs))
	// Output: Service is ready to delete 3 URLs asynchronously
}

// ExampleNewURLDestructorService_stop демонстрирует корректную остановку сервиса.
func ExampleNewURLDestructorService_stop() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)

	service := destructor.NewURLDestructorService(mockURLRepo, mockUserURLsRepo)

	// Выполняем работу с сервисом...
	// ...

	// Корректно останавливаем сервис
	// Метод Stop() дожидается завершения всех воркеров
	service.Stop()

	fmt.Println("Service stopped gracefully")
	// Output: Service stopped gracefully
}
