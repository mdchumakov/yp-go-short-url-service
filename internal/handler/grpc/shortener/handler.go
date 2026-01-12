package shortener

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"yp-go-short-url-service/internal/config"
	pb "yp-go-short-url-service/internal/generated/api/proto"
	"yp-go-short-url-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pb.UnimplementedShortenerServiceServer
	service service.URLShortenerService
	baseURL string
}

func NewHandler(
	service service.URLShortenerService,
	settings *config.Settings,
) *Handler {
	return &Handler{
		service: service,
		baseURL: settings.GetBaseURL(),
	}
}

func (h *Handler) ShortenURL(
	ctx context.Context,
	req *pb.URLShortenRequest,
) (*pb.URLShortenResponse, error) {
	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	shortURL, err := h.service.ShortURL(ctx, req.GetUrl())
	if err != nil {
		if service.IsAlreadyExistsError(err) && shortURL != "" {
			resultURL := h.buildShortURL(shortURL)
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

	resultURL := h.buildShortURL(shortURL)
	return pb.URLShortenResponse_builder{
		Result:     resultURL,
		StatusCode: http.StatusCreated,
	}.Build(), nil
}

func (h *Handler) buildShortURL(shortedURL string) string {
	return fmt.Sprintf("%s/%s", strings.TrimRight(h.baseURL, "/"), shortedURL)
}
