package stats

import (
	"context"
	"yp-go-short-url-service/internal/repository"
	"yp-go-short-url-service/internal/service"
)

func New(
	userRepo repository.UserRepositoryReader,
	urlsRepo repository.URLRepositoryReader,
) service.StatsService {
	return &serviceImpl{
		userRepo: userRepo,
		urlsRepo: urlsRepo,
	}
}

type serviceImpl struct {
	userRepo repository.UserRepositoryReader
	urlsRepo repository.URLRepositoryReader
}

func (s *serviceImpl) GetTotalURLsCount(ctx context.Context) (int64, error) {
	return s.urlsRepo.GetTotalCount(ctx)
}

func (s *serviceImpl) GetTotalUsersCount(ctx context.Context) (int64, error) {
	return s.userRepo.GetUsersCount(ctx)
}
