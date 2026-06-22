package handler

import (
	"log"
	"net/http"

	"github.com/cyancyan2020/iam-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logService service.LogService
}

func NewLogHandler(logService service.LogService) *LogHandler {
	return &LogHandler{logService: logService}
}

func (h *LogHandler) Query(c *gin.Context) {
	var query service.LogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}

	result, err := h.logService.Query(c.Request.Context(), &query)
	if err != nil {
		log.Printf("日志查询 error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务器内部错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}
