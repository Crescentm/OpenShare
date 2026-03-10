package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openshare/backend/pkg/logger"
)

// Logger 日志中间件
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		if status >= 500 {
			log.Error("HTTP Request",
				"status", status,
				"method", method,
				"path", path,
				"ip", clientIP,
				"latency", latency,
				"error", c.Errors.String(),
			)
		} else if status >= 400 {
			log.Warn("HTTP Request",
				"status", status,
				"method", method,
				"path", path,
				"ip", clientIP,
				"latency", latency,
			)
		} else {
			log.Info("HTTP Request",
				"status", status,
				"method", method,
				"path", path,
				"ip", clientIP,
				"latency", latency,
			)
		}
	}
}
