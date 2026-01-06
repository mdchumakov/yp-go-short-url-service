package stats

import (
	"net/http"
	handler_ "yp-go-short-url-service/internal/handler"
	"yp-go-short-url-service/internal/middleware"
	"yp-go-short-url-service/internal/service"

	"github.com/gin-gonic/gin"
)

type handler struct {
	trustedSubnet string
	service       service.StatsService
}

func New(service service.StatsService, trustedSubnet string) handler_.Handler {
	return &handler{
		service:       service,
		trustedSubnet: trustedSubnet,
	}
}

func (h *handler) Handle(c *gin.Context) {
	logger := middleware.GetLogger(c.Request.Context())
	requestID := middleware.ExtractRequestID(c.Request.Context())

	usersCount, err := h.service.GetTotalUsersCount(c.Request.Context())
	if err != nil {
		logger.Errorw("failed to get users count", "error", err, "id", requestID)
		c.String(http.StatusInternalServerError, "failed to get users count")
		return
	}

	urlsCount, err := h.service.GetTotalURLsCount(c.Request.Context())
	if err != nil {
		logger.Errorw("failed to get URLs count", "error", err, "id", requestID)
		c.String(http.StatusInternalServerError, "failed to get URLs count")
		return
	}

	resp := Response{
		URLsCount:  urlsCount,
		UsersCount: usersCount,
	}
	logger.Infow("responding to request", "id", requestID, "response", resp)
	c.JSON(http.StatusOK, resp)
}
