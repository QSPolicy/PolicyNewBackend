package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response 统一返回结构
type Response struct {
	Code    int         `json:"code"`           // 业务状态码
	Message string      `json:"message"`        // 提示信息
	Data    interface{} `json:"data,omitempty"` // 数据
}

// Success 成功返回
func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// Fail 失败返回 (业务错误)
func Fail(c echo.Context, code int, msg string) error {
	return c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
		Data:    nil,
	})
}

// Error 错误返回 (系统错误) 此时状态码为 httpCode
func Error(c echo.Context, httpCode int, msg string) error {
	return c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: msg,
		Data:    nil,
	})
}
