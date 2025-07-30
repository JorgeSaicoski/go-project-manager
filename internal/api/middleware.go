package api

import (
	"github.com/JorgeSaicoski/microservice-commons/middleware"
	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return middleware.DefaultLoggingMiddleware()
}
