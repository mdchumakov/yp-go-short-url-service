package grpc

import (
	"context"
	"net/http"
	pb "yp-go-short-url-service/internal/generated/api/proto"
	"yp-go-short-url-service/internal/middleware"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *RPCService) ListUserURLs(
	ctx context.Context,
	_ *emptypb.Empty,
) (*pb.UserURLsResponse, error) {
	user := middleware.GetJWTUserFromContext(ctx)
	if user == nil {
		return pb.UserURLsResponse_builder{
			StatusCode: http.StatusUnauthorized,
			Error:      &[]string{"user not found"}[0],
		}.Build(), status.Error(codes.Unauthenticated, "user not found")
	}

	userURLs, err := s.deps.extractorService.ExtractUserURLs(ctx, user.ID)
	if err != nil {
		return pb.UserURLsResponse_builder{
			StatusCode: http.StatusInternalServerError,
			Error:      &[]string{"failed to extract user URLs"}[0],
		}.Build(), status.Error(codes.Internal, err.Error())
	}

	if len(userURLs) == 0 {
		return pb.UserURLsResponse_builder{
			Url:        []*pb.URLData{},
			StatusCode: http.StatusNoContent,
		}.Build(), nil
	}

	// Преобразуем данные
	urls := make([]*pb.URLData, len(userURLs))
	for i, url := range userURLs {
		urls[i] = pb.URLData_builder{
			ShortUrl:    s.buildShortURL(url.ShortURL),
			OriginalUrl: url.LongURL,
		}.Build()
	}

	return pb.UserURLsResponse_builder{
		Url:        urls,
		StatusCode: http.StatusOK,
	}.Build(), nil
}
