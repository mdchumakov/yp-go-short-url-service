package destructor

import (
	"context"
	"errors"
	"sync"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
)

func NewURLDestructorService(
	urlRepository repository.URLRepository,
	userURLsRepository repository.UserURLsRepository,
) service.URLDestructorService {
	destructorService := &urlDestructorService{
		urlRepository:      urlRepository,
		userURLsRepository: userURLsRepository,
		deleteChan:         make(chan deleteRequest, 100), // буфер для 100 запросов
		stopChan:           make(chan struct{}),
		wg:                 &sync.WaitGroup{},
	}

	// Запускаем горутину для обработки удалений
	destructorService.wg.Add(1)
	go destructorService.deleteWorker()

	return destructorService
}

type deleteRequest struct {
	shortURLs []string
	userID    string
	ctx       context.Context
}

type urlDestructorService struct {
	urlRepository      repository.URLRepository
	userURLsRepository repository.UserURLsRepository
	deleteChan         chan deleteRequest
	stopChan           chan struct{}
	wg                 *sync.WaitGroup
}

// deleteWorker - горутина для асинхронной обработки удалений
func (s *urlDestructorService) deleteWorker() {
	defer s.wg.Done()

	for {
		select {
		case req := <-s.deleteChan:
			// Выполняем удаление в отдельной горутине
			err := s.userURLsRepository.DeleteURLsWithUser(req.ctx, req.shortURLs, req.userID)
			if err != nil {
				logger := middleware.GetLogger(req.ctx)
				requestID := middleware.ExtractRequestID(req.ctx)
				logger.Errorw("Failed to delete URLs from storage in async worker",
					"error", err,
					"request_id", requestID,
					"short_urls", req.shortURLs,
					"user_id", req.userID,
				)
			}
		case <-s.stopChan:
			// Получаем сигнал остановки
			return
		}
	}
}

// Stop - метод для корректной остановки сервиса
func (s *urlDestructorService) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

func (s *urlDestructorService) DeleteURL(ctx context.Context, shortURL string) error {
	//TODO implement me
	panic("implement me")
}

func (s *urlDestructorService) DeleteURLsByBatch(ctx context.Context, shortURLs []string) error {
	logger := middleware.GetLogger(ctx)
	requestID := middleware.ExtractRequestID(ctx)

	user := middleware.GetJWTUserFromContext(ctx)
	if user == nil {
		return errors.New("user is not authenticated")
	}

	logger.Debugw("Starting async URL deletion process",
		"request_id", requestID,
		"short_urls_count", len(shortURLs),
	)

	// Отправляем запрос на удаление в канал для асинхронной обработки
	select {
	case s.deleteChan <- deleteRequest{
		shortURLs: shortURLs,
		userID:    user.ID,
		ctx:       ctx,
	}:
		logger.Debugw("URL deletion request sent to async worker",
			"request_id", requestID,
		)
		return nil
	default:
		// Если канал переполнен, возвращаем ошибку
		logger.Errorw("Delete channel is full, cannot process deletion request",
			"request_id", requestID,
		)
		return errors.New("delete service is overloaded, try again later")
	}
}
