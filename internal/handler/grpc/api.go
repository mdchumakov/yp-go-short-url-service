package grpc

import (
	"yp-go-short-url-service/internal/config"
	pb "yp-go-short-url-service/internal/generated/api/proto"
	"yp-go-short-url-service/internal/service"
)

type RPCService struct {
	pb.UnimplementedShortenerServiceServer
	deps dependencies
}

type dependencies struct {
	shortenerService service.URLShortenerService
	extractorService service.URLExtractorService
	baseURL          string
}

func NewRPCService(
	shortenerService service.URLShortenerService,
	extractorService service.URLExtractorService,
	settings *config.Settings,
) *RPCService {
	deps := dependencies{
		shortenerService: shortenerService,
		extractorService: extractorService,
		baseURL:          settings.GetBaseURL(),
	}
	return &RPCService{
		deps: deps,
	}
}
