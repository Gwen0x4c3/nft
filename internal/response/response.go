package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nft-platform/internal/errors"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}

// ErrorInfo represents error information in response
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// MetaInfo represents pagination and other metadata
type MetaInfo struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
	HasNext    bool `json:"has_next,omitempty"`
	HasPrev    bool `json:"has_prev,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *gin.Context, statusCode int, data interface{}, meta *MetaInfo) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Error sends an error response
func Error(c *gin.Context, err error) {
	if apiErr, ok := err.(*errors.APIError); ok {
		c.JSON(apiErr.HTTPStatus, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    string(apiErr.Code),
				Message: apiErr.Message,
				Details: apiErr.Details,
			},
		})
		return
	}

	// Fallback for non-API errors
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(errors.ErrInternal),
			Message: "Internal server error",
		},
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	Success(c, http.StatusCreated, data)
}

// OK sends a 200 OK response
func OK(c *gin.Context, data interface{}) {
	Success(c, http.StatusOK, data)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string, details map[string]interface{}) {
	apiErr := errors.NewAPIErrorWithDetails(errors.ErrInvalidRequest, message, details)
	Error(c, apiErr)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	apiErr := errors.NewAPIError(errors.ErrUnauthorized, message)
	Error(c, apiErr)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	apiErr := errors.NewAPIError(errors.ErrForbidden, message)
	Error(c, apiErr)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, resource string) {
	apiErr := errors.NewNotFoundError(resource)
	Error(c, apiErr)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	apiErr := errors.NewAPIError(errors.ErrAlreadyExists, message)
	Error(c, apiErr)
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(c *gin.Context, message string) {
	apiErr := errors.NewAPIError(errors.ErrResourceExhausted, message)
	Error(c, apiErr)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	apiErr := errors.NewAPIError(errors.ErrInternal, message)
	Error(c, apiErr)
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string, validationErrors map[string]interface{}) {
	apiErr := errors.NewValidationError(message, validationErrors)
	Error(c, apiErr)
}

// PaginationMeta creates pagination metadata
func PaginationMeta(page, limit int, total int64) *MetaInfo {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	return &MetaInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// ListResponse creates a standard list response with pagination
func ListResponse(c *gin.Context, data interface{}, page, limit int, total int64) {
	meta := PaginationMeta(page, limit, total)
	SuccessWithMeta(c, http.StatusOK, data, meta)
}