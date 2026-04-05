package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const (
	CodeSuccess = 0
	CodeError   = 1
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func ErrorWithStatus(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    CodeError,
		Message: message,
		Data:    nil,
	})
}

func BadRequest(c *gin.Context, message string) {
	ErrorWithStatus(c, http.StatusBadRequest, message)
}

func InternalError(c *gin.Context, message string) {
	ErrorWithStatus(c, http.StatusInternalServerError, message)
}

func NotFound(c *gin.Context, message string) {
	ErrorWithStatus(c, http.StatusNotFound, message)
}

func Unauthorized(c *gin.Context, message string) {
	ErrorWithStatus(c, http.StatusUnauthorized, message)
}
