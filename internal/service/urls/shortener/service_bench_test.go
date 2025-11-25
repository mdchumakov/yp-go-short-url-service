package shortener

import (
	"context"
	"testing"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/model"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/repository/mock"

	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// setupBenchmarkContext создает контекст с логгером для бенчмарков
func setupBenchmarkContext() context.Context {
	logger, _ := zap.NewDevelopment()
	return middleware.WithLogger(context.Background(), logger.Sugar())
}

// BenchmarkShortURL_NewURL бенчмарк для создания нового короткого URL
func BenchmarkShortURL_NewURL(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)

	service := NewURLShortenerService(mockRepo, mockUserURLsRepo, nil)
	ctx := setupBenchmarkContext()

	longURL := "https://example.com/very/long/url/path/that/needs/to/be/shortened"

	// Настраиваем моки для каждого итерации
	mockRepo.EXPECT().
		GetByLongURL(gomock.Any(), longURL).
		Return(nil, repository.ErrURLNotFound).
		Times(b.N)

	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ShortURL(ctx, longURL)
	}
}

// BenchmarkShortURL_ExistingURL бенчмарк для случая, когда URL уже существует
func BenchmarkShortURL_ExistingURL(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)

	service := NewURLShortenerService(mockRepo, mockUserURLsRepo, nil)
	ctx := setupBenchmarkContext()

	longURL := "https://example.com/existing/url"
	existingURL := &model.URLsModel{
		ShortURL: "abc12345",
		LongURL:  longURL,
	}

	mockRepo.EXPECT().
		GetByLongURL(gomock.Any(), longURL).
		Return(existingURL, nil).
		Times(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ShortURL(ctx, longURL)
	}
}

// BenchmarkShortURLsByBatch бенчмарк для пакетного создания URL
func BenchmarkShortURLsByBatch(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)

	service := NewURLShortenerService(mockRepo, mockUserURLsRepo, nil)
	ctx := setupBenchmarkContext()

	batchSize := 10
	longURLs := make([]map[string]string, batchSize)
	for i := 0; i < batchSize; i++ {
		longURLs[i] = map[string]string{
			"correlation_id": string(rune(i)),
			"original_url":   "https://example.com/url" + string(rune(i)),
		}
	}

	// Настраиваем моки для каждой итерации
	for i := 0; i < b.N; i++ {
		for _, urlData := range longURLs {
			mockRepo.EXPECT().
				GetByLongURL(gomock.Any(), urlData["original_url"]).
				Return(nil, repository.ErrURLNotFound)
		}
		mockRepo.EXPECT().
			CreateBatch(gomock.Any(), gomock.Any()).
			Return(nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ShortURLsByBatch(ctx, longURLs)
	}
}

// BenchmarkShortURLsByBatch_Large бенчмарк для большого пакета URL
func BenchmarkShortURLsByBatch_Large(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mock.NewMockURLRepository(ctrl)
	mockUserURLsRepo := mock.NewMockUserURLsRepository(ctrl)

	service := NewURLShortenerService(mockRepo, mockUserURLsRepo, nil)
	ctx := setupBenchmarkContext()

	batchSize := 100
	longURLs := make([]map[string]string, batchSize)
	for i := 0; i < batchSize; i++ {
		longURLs[i] = map[string]string{
			"correlation_id": string(rune(i)),
			"original_url":   "https://example.com/url" + string(rune(i)),
		}
	}

	// Настраиваем моки для каждой итерации
	for i := 0; i < b.N; i++ {
		for _, urlData := range longURLs {
			mockRepo.EXPECT().
				GetByLongURL(gomock.Any(), urlData["original_url"]).
				Return(nil, repository.ErrURLNotFound)
		}
		mockRepo.EXPECT().
			CreateBatch(gomock.Any(), gomock.Any()).
			Return(nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ShortURLsByBatch(ctx, longURLs)
	}
}
