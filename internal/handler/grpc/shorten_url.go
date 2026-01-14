package grpc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	pb "yp-go-short-url-service/internal/generated/api/proto"
	"yp-go-short-url-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RPCService) ShortenURL(
	ctx context.Context,
	req *pb.URLShortenRequest,
) (*pb.URLShortenResponse, error) {
	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	shortURL, err := s.deps.shortenerService.ShortURL(ctx, req.GetUrl())
	if err != nil {
		if service.IsAlreadyExistsError(err) && shortURL != "" {
			resultURL := s.buildShortURL(shortURL)
			return pb.URLShortenResponse_builder{
				Result:     resultURL,
				StatusCode: http.StatusConflict,
			}.Build(), nil
		}

		return pb.URLShortenResponse_builder{
			Result:     "",
			StatusCode: http.StatusInternalServerError,
			Error:      &[]string{err.Error()}[0],
		}.Build(), status.Error(codes.Internal, err.Error())
	}

	resultURL := s.buildShortURL(shortURL)
	return pb.URLShortenResponse_builder{
		Result:     resultURL,
		StatusCode: http.StatusCreated,
	}.Build(), nil
}

func (s *RPCService) buildShortURL(shortedURL string) string {
	return fmt.Sprintf("%s/%s", strings.TrimRight(s.deps.baseURL, "/"), shortedURL)
}
