package destructor

import (
	"context"
	"errors"
	"sync"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"

	"go.uber.org/zap"
)

const (
	// Количество воркеров для обработки удалений
	numWorkers = 3
	// Размер буфера канала
	channelBufferSize = 100
)

func NewURLDestructorService(
	urlRepository repository.URLRepository,
	userURLsRepository repository.UserURLsRepository,
) service.URLDestructorService {
	destructorService := &urlDestructorService{
		urlRepository:      urlRepository,
		userURLsRepository: userURLsRepository,
		deleteChan:         make(chan deleteRequest, channelBufferSize),
		stopChan:           make(chan struct{}),
		wg:                 &sync.WaitGroup{},
	}

	// Запускаем несколько горутин для обработки удалений (паттерн fan-in)
	for i := 0; i < numWorkers; i++ {
		destructorService.wg.Add(1)
		go destructorService.deleteWorker(i)
	}

	return destructorService
}

type deleteRequest struct {
	shortURLs []string
	userID    string
	requestID string
	logger    *zap.SugaredLogger
}

type urlDestructorService struct {
	urlRepository      repository.URLRepository
	userURLsRepository repository.UserURLsRepository
	deleteChan         chan deleteRequest
	stopChan           chan struct{}
	wg                 *sync.WaitGroup
}

// deleteWorker - горутина для асинхронной обработки удалений (паттерн fan-in)
func (s *urlDestructorService) deleteWorker(workerID int) {
	defer s.wg.Done()

	for {
		select {
		case req := <-s.deleteChan:
			// Создаем новый контекст для асинхронной операции
			ctx := context.Background()

			req.logger.Debugw("Worker started processing delete request",
				"worker_id", workerID,
				"request_id", req.requestID,
				"short_urls_count", len(req.shortURLs),
			)

			// Выполняем удаление в отдельной горутине
			err := s.userURLsRepository.DeleteURLsWithUser(ctx, req.shortURLs, req.userID)
			if err != nil {
				req.logger.Errorw("Failed to delete URLs from storage in async worker",
					"error", err,
					"worker_id", workerID,
					"request_id", req.requestID,
					"short_urls", req.shortURLs,
					"user_id", req.userID,
				)
			} else {
				req.logger.Debugw("Successfully deleted URLs in worker",
					"worker_id", workerID,
					"request_id", req.requestID,
					"short_urls_count", len(req.shortURLs),
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
		"num_workers", numWorkers,
	)

	// Отправляем запрос на удаление в канал для асинхронной обработки
	select {
	case s.deleteChan <- deleteRequest{
		shortURLs: shortURLs,
		userID:    user.ID,
		requestID: requestID,
		logger:    logger,
	}:
		logger.Debugw("URL deletion request sent to async worker pool",
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
