package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/openshare/backend/pkg/logger"
	"github.com/openshare/backend/pkg/response"
)

// Recovery 异常恢复中间件
func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("Panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)

				response.Error(c, http.StatusInternalServerError, response.CodeInternalError, "internal server error")
				c.Abort()
			}
		}()

		c.Next()
	}
}
