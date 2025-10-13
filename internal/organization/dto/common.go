package dto

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageRequest 分页请求
type PageRequest struct {
	Page     int `json:"page" form:"page"`         // 页码，从 1 开始
	PageSize int `json:"page_size" form:"page_size"` // 每页数量
}

// PageResponse 分页响应
type PageResponse struct {
	Total    int64       `json:"total"`     // 总数
	Page     int         `json:"page"`      // 当前页
	PageSize int         `json:"page_size"` // 每页数量
	Data     interface{} `json:"data"`      // 数据列表
}

// IDRequest ID 请求
type IDRequest struct {
	ID string `json:"id" uri:"id" binding:"required,uuid"`
}

// Success 成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// Error 错误响应
func Error(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// PageSuccess 分页成功响应
func PageSuccess(total int64, page, pageSize int, data interface{}) *Response {
	return &Response{
		Code:    0,
		Message: "success",
		Data: &PageResponse{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			Data:     data,
		},
	}
}
