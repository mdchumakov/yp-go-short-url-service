package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"yp-go-short-url-service/internal/model"
)

type ExtractingFullLink struct {
	db *gorm.DB
}

func NewExtractingFullLink(db *gorm.DB) *ExtractingFullLink {
	return &ExtractingFullLink{db: db}
}

func (h *ExtractingFullLink) Handle(c *gin.Context) {
	var urlModel model.URL

	shortURL := c.Param("shortURL")

	h.db.First(&urlModel, "short_url = ?", shortURL)
	if urlModel.ID == 0 {
		c.String(http.StatusBadRequest, "Ссылка не найдена")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, urlModel.LongURL)
}
