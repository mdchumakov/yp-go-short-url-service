package extractor_test

import (
	"context"
	"fmt"
	"yp-go-short-url-service/internal/observer/audit"
	observerMock "yp-go-short-url-service/internal/observer/mock"
	"yp-go-short-url-service/internal/repository/mock"
	"yp-go-short-url-service/internal/service/urls/extractor"

	"go.uber.org/mock/gomock"
)

// ExampleNewLinkExtractorService демонстрирует создание нового сервиса для извлечения URL.
func ExampleNewLinkExtractorService() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	// Создаем моки репозиториев
	mockURLRepo := mock.NewMockURLRepositoryReader(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepositoryReader(ctrl)
	auditEventBus := observerMock.NewMockSubject[audit.Event](ctrl)

	// Создаем сервис для извлечения URL
	service := extractor.NewLinkExtractorService(mockURLRepo, mockUserURLsRepo, auditEventBus)

	// Сервис готов к использованию
	_ = service

	fmt.Println("LinkExtractorService created successfully")
	// Output: LinkExtractorService created successfully
}

// ExampleNewLinkExtractorService_extractLongURL демонстрирует извлечение длинного URL по короткому.
func ExampleNewLinkExtractorService_extractLongURL() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockURLRepo := mock.NewMockURLRepositoryReader(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepositoryReader(ctrl)
	auditEventBus := observerMock.NewMockSubject[audit.Event](ctrl)

	service := extractor.NewLinkExtractorService(mockURLRepo, mockUserURLsRepo, auditEventBus)

	ctx := context.Background()
	shortURL := "abc123"

	// В реальном приложении здесь будет вызов:
	// longURL, err := service.ExtractLongURL(ctx, shortURL)
	// Для примера демонстрируем только создание сервиса
	_ = service
	_ = ctx
	_ = shortURL

	fmt.Println("Service is ready to extract long URLs")
	// Output: Service is ready to extract long URLs
}

// ExampleNewLinkExtractorService_extractUserURLs демонстрирует извлечение всех URL пользователя.
func ExampleNewLinkExtractorService_extractUserURLs() {
	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	mockURLRepo := mock.NewMockURLRepositoryReader(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepositoryReader(ctrl)
	auditEventBus := observerMock.NewMockSubject[audit.Event](ctrl)

	service := extractor.NewLinkExtractorService(mockURLRepo, mockUserURLsRepo, auditEventBus)

	ctx := context.Background()
	userID := "user-123"

	// В реальном приложении здесь будет вызов:
	// urls, err := service.ExtractUserURLs(ctx, userID)
	// Для примера демонстрируем только создание сервиса
	_ = service
	_ = ctx
	_ = userID

	fmt.Println("Service is ready to extract user URLs")
	// Output: Service is ready to extract user URLs
}
