package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openshare/backend/pkg/response"
)

// Health 健康检查接口
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "openshare",
	})
}

// NotImplemented 未实现的接口占位
func NotImplemented(c *gin.Context) {
	response.Error(c, http.StatusNotImplemented, 501, "not implemented")
}
