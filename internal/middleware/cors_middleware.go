package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware устанавливает заголовки CORS для обработки кросс-доменных запросов.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Разрешаем запросы со всех доменов
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204) // Для OPTIONS-запросов (preflight) возвращаем 204
			return
		}

		c.Next()
	}
}