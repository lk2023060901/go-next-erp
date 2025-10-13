package dto

// ===========================
// 通用响应结构
// ===========================

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponseWithCode 带状态码的错误响应
func ErrorResponseWithCode(code int, message string, err error) *ErrorResponse {
	resp := &ErrorResponse{
		Code:    code,
		Message: message,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

// BadRequestResponse 400错误
func BadRequestResponse(message string, err error) *ErrorResponse {
	return ErrorResponseWithCode(400, message, err)
}

// UnauthorizedResponse 401错误
func UnauthorizedResponse(message string) *ErrorResponse {
	return ErrorResponseWithCode(401, message, nil)
}

// ForbiddenResponse 403错误
func ForbiddenResponse(message string) *ErrorResponse {
	return ErrorResponseWithCode(403, message, nil)
}

// NotFoundResponse 404错误
func NotFoundResponse(message string) *ErrorResponse {
	return ErrorResponseWithCode(404, message, nil)
}

// ConflictResponse 409错误
func ConflictResponse(message string, err error) *ErrorResponse {
	return ErrorResponseWithCode(409, message, err)
}

// InternalErrorResponse 500错误
func InternalErrorResponse(err error) *ErrorResponse {
	return ErrorResponseWithCode(500, "Internal server error", err)
}

// ===========================
// 分页相关
// ===========================

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// CalculatePagination 计算分页信息
func CalculatePagination(page, pageSize int, total int64) *PaginationMeta {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

// GetOffset 计算偏移量
func (p *PaginationMeta) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *PaginationMeta) GetLimit() int {
	return p.PageSize
}
