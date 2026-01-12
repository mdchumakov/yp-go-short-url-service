package grpc

import (
	"context"
	"net/http"
	pb "yp-go-short-url-service/internal/generated/api/proto"
	"yp-go-short-url-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *RPCService) ExpandURL(
	ctx context.Context,
	req *pb.URLExpandRequest,
) (*pb.URLExpandResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	longURL, err := s.deps.extractorService.ExtractLongURL(ctx, req.GetId())
	if err != nil {
		if service.IsDeletedError(err) {
			return pb.URLExpandResponse_builder{
				StatusCode: http.StatusGone,
				Error:      &[]string{"Ссылка была удалена"}[0],
			}.Build(), nil
		}
		return pb.URLExpandResponse_builder{
			StatusCode: http.StatusInternalServerError,
			Error:      &[]string{err.Error()}[0],
		}.Build(), status.Error(codes.Internal, err.Error())
	}

	if longURL == "" {
		return pb.URLExpandResponse_builder{
			StatusCode: http.StatusBadRequest,
			Error:      &[]string{"Ссылка не найдена"}[0],
		}.Build(), nil
	}

	return pb.URLExpandResponse_builder{
		Result:     longURL,
		StatusCode: http.StatusTemporaryRedirect,
	}.Build(), nil
}
