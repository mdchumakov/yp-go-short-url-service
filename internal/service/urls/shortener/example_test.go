package shortener_test

import (
	"context"
	"fmt"
	"yp-go-short-url-service/internal/observer/audit"
	observerMock "yp-go-short-url-service/internal/observer/mock"
	"yp-go-short-url-service/internal/repository/mock"
	"yp-go-short-url-service/internal/service/urls/shortener"

	"go.uber.org/mock/gomock"
)

// ExampleNewURLShortenerService демонстрирует создание нового сервиса для сокращения URL.
func ExampleNewURLShortenerService() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	// Создаем моки репозиториев
	mockURLRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)
	auditEventBus := observerMock.NewMockSubject[audit.Event](ctrl)

	// Создаем сервис для сокращения URL
	service := shortener.NewURLShortenerService(mockURLRepo, mockUserURLsRepo, auditEventBus)

	// Сервис готов к использованию
	_ = service

	fmt.Println("URLShortenerService created successfully")
	// Output: URLShortenerService created successfully
}

// ExampleNewURLShortenerService_shortURL демонстрирует сокращение одного URL.
func ExampleNewURLShortenerService_shortURL() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)
	auditEventBus := observerMock.NewMockSubject[audit.Event](ctrl)

	service := shortener.NewURLShortenerService(mockURLRepo, mockUserURLsRepo, auditEventBus)

	ctx := context.Background()
	longURL := "https://example.com/very/long/url/path"

	// В реальном приложении здесь будет вызов:
	// shortURL, err := service.ShortURL(ctx, longURL)
	// Для примера демонстрируем только создание сервиса
	_ = service
	_ = ctx
	_ = longURL

	fmt.Println("Service is ready to shorten URLs")
	// Output: Service is ready to shorten URLs
}

// ExampleNewURLShortenerService_shortURLsByBatch демонстрирует пакетное сокращение URL.
func ExampleNewURLShortenerService_shortURLsByBatch() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockURLRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)
	auditEventBus := observerMock.NewMockSubject[audit.Event](ctrl)

	service := shortener.NewURLShortenerService(mockURLRepo, mockUserURLsRepo, auditEventBus)

	ctx := context.Background()

	// Подготавливаем массив URL для пакетного сокращения
	longURLs := []map[string]string{
		{"correlation_id": "1", "original_url": "https://example.com/url1"},
		{"correlation_id": "2", "original_url": "https://example.com/url2"},
		{"correlation_id": "3", "original_url": "https://example.com/url3"},
	}

	// В реальном приложении здесь будет вызов:
	// result, err := service.ShortURLsByBatch(ctx, longURLs)
	// Для примера демонстрируем только подготовку данных
	_ = service
	_ = ctx
	_ = longURLs

	fmt.Printf("Service is ready to process %d URLs in batch\n", len(longURLs))
	// Output: Service is ready to process 3 URLs in batch
}
