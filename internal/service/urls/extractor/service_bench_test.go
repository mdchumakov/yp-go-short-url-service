package extractor

import (
	"context"
	"testing"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository/mock"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// setupBenchmarkContext создает контекст с логгером для бенчмарков
func setupBenchmarkContext() context.Context {
	logger, _ := zap.NewDevelopment()
	return middleware.WithLogger(context.Background(), logger.Sugar())
}

// BenchmarkExtractLongURL бенчмарк для извлечения длинного URL по короткому
func BenchmarkExtractLongURL(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mock.NewMockURLRepositoryReader(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepositoryReader(ctrl)

	service := NewLinkExtractorService(mockRepo, mockUserURLsRepo, nil)
	ctx := setupBenchmarkContext()

	shortURL := "abc12345"
	longURL := &model.URLsModel{
		ShortURL:  shortURL,
		LongURL:   "https://example.com/very/long/url/path",
		IsDeleted: false,
	}

	mockRepo.EXPECT().
		GetByShortURL(gomock.Any(), shortURL).
		Return(longURL, nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ExtractLongURL(ctx, shortURL)
	}
}

// BenchmarkExtractUserURLs бенчмарк для извлечения URL пользователя
func BenchmarkExtractUserURLs(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mock.NewMockURLRepositoryReader(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepositoryReader(ctrl)

	service := NewLinkExtractorService(mockRepo, mockUserURLsRepo, nil)
	ctx := setupBenchmarkContext()

	userID := "user123"
	urls := make([]*model.URLsModel, 10)
	for i := 0; i < 10; i++ {
		urls[i] = &model.URLsModel{
			ShortURL: "short" + string(rune(i)),
			LongURL:  "https://example.com/url" + string(rune(i)),
		}
	}

	mockUserURLsRepo.EXPECT().
		GetByUserID(gomock.Any(), userID).
		Return(urls, nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ExtractUserURLs(ctx, userID)
	}
}
