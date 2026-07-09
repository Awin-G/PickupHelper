package handler

import (
	stderrors "errors"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
)

// Response is the unified JSON envelope returned by every API endpoint.
// See 详细设计文档/api详细设计.md 1.3.
type Response struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    any    `json:"data,omitempty"`
	TraceID string `json:"trace_id"`
}

// PagedData wraps list endpoints with pagination metadata.
// See 详细设计文档/api详细设计.md 1.4.
type PagedData struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// Success writes a 200 response with code=0 and the given data payload.
func Success(c *gin.Context, data any) {
	c.JSON(200, Response{
		Code:    apperrors.Success,
		Msg:     "success",
		Data:    data,
		TraceID: log.TraceID(c.Request.Context()),
	})
}

// SuccessPaged writes a 200 response with a PagedData payload.
func SuccessPaged(c *gin.Context, list any, total int64, page, pageSize int) {
	c.JSON(200, Response{
		Code: apperrors.Success,
		Msg:  "success",
		Data: PagedData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		TraceID: log.TraceID(c.Request.Context()),
	})
}

// Error writes the appropriate HTTP status and unified Response body for err.
// If err is (or wraps) an *apperrors.AppError, its Code/HTTPStatus/Msg are used.
// Otherwise err is treated as an internal error (code=10009, HTTP 500).
func Error(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if stderrors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, Response{
			Code:    appErr.Code,
			Msg:     appErr.Msg,
			TraceID: log.TraceID(c.Request.Context()),
		})
		return
	}
	// Generic error → wrap as internal.
	c.JSON(500, Response{
		Code:    apperrors.ErrInternal,
		Msg:     err.Error(),
		TraceID: log.TraceID(c.Request.Context()),
	})
}
