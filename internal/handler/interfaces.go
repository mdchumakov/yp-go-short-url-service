package handler

import "github.com/gin-gonic/gin"

// Handler определяет интерфейс для обработчиков HTTP-запросов.
// Все обработчики должны реализовывать метод Handle для обработки запросов через Gin контекст.
type Handler interface {
	Handle(c *gin.Context)
}
