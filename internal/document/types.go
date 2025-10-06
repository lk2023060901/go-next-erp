package document

// TaskState 任务状态
type TaskState string

const (
	TaskStatePending TaskState = "pending" // 排队中
	TaskStateRunning TaskState = "running" // 运行中
	TaskStateDone    TaskState = "done"    // 完成
	TaskStateFailed  TaskState = "failed"  // 失败
)

// CreateTaskRequest 创建单个解析任务请求
type CreateTaskRequest struct {
	URL            string   `json:"url"`                       // 文件 URL (必填)
	IsOCR          bool     `json:"is_ocr,omitempty"`          // 是否启动 OCR，默认 false
	EnableFormula  bool     `json:"enable_formula,omitempty"`  // 是否开启公式识别，默认 true
	EnableTable    bool     `json:"enable_table,omitempty"`    // 是否开启表格识别，默认 true
	Language       string   `json:"language,omitempty"`        // 文档语言，默认 ch
	DataID         string   `json:"data_id,omitempty"`         // 数据 ID
	Callback       string   `json:"callback,omitempty"`        // 回调通知 URL
	Seed           string   `json:"seed,omitempty"`            // 随机字符串
	ExtraFormats   []string `json:"extra_formats,omitempty"`   // 额外导出格式: docx, html, latex
	PageRanges     string   `json:"page_ranges,omitempty"`     // 页码范围，例如: "1-600"
	ModelVersion   string   `json:"model_version,omitempty"`   // 模型版本: pipeline 或 vlm，默认 pipeline
}

// CreateTaskResponse 创建任务响应
type CreateTaskResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
	TraceID string `json:"trace_id"`
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID     string    `json:"task_id"`
	State      TaskState `json:"state"`
	FullZipURL string    `json:"full_zip_url,omitempty"` // 完成时提供
	ErrorMsg   string    `json:"err_msg,omitempty"`
}

// GetTaskResultResponse 获取任务结果响应
type GetTaskResultResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"msg"`
	Data    TaskResult `json:"data"`
	TraceID string     `json:"trace_id"`
}

// BatchFileItem 批量文件项
type BatchFileItem struct {
	Name  string `json:"name"`            // 文件名 (必填)
	IsOCR bool   `json:"is_ocr"`          // 是否启动 OCR
	DataID string `json:"data_id,omitempty"` // 数据 ID
}

// BatchURLItem 批量 URL 项
type BatchURLItem struct {
	URL    string `json:"url"`             // 文件 URL (必填)
	IsOCR  bool   `json:"is_ocr"`          // 是否启动 OCR
	DataID string `json:"data_id,omitempty"` // 数据 ID
}

// CreateBatchUploadRequest 批量上传解析请求
type CreateBatchUploadRequest struct {
	EnableFormula bool            `json:"enable_formula"` // 是否开启公式识别
	Language      string          `json:"language"`       // 文档语言
	EnableTable   bool            `json:"enable_table"`   // 是否开启表格识别
	Files         []BatchFileItem `json:"files"`          // 文件列表
}

// CreateBatchUploadResponse 批量上传响应
type CreateBatchUploadResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    struct {
		BatchID  string   `json:"batch_id"`
		FileURLs []string `json:"file_urls"` // 上传 URL 列表
	} `json:"data"`
	TraceID string `json:"trace_id"`
}

// CreateBatchURLRequest 批量 URL 解析请求
type CreateBatchURLRequest struct {
	EnableFormula bool           `json:"enable_formula"` // 是否开启公式识别
	Language      string         `json:"language"`       // 文档语言
	EnableTable   bool           `json:"enable_table"`   // 是否开启表格识别
	Files         []BatchURLItem `json:"files"`          // URL 列表
}

// CreateBatchURLResponse 批量 URL 响应
type CreateBatchURLResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    struct {
		BatchID string `json:"batch_id"`
	} `json:"data"`
	TraceID string `json:"trace_id"`
}

// ExtractProgress 解析进度
type ExtractProgress struct {
	ExtractedPages int    `json:"extracted_pages"` // 已解析页数
	TotalPages     int    `json:"total_pages"`     // 总页数
	StartTime      string `json:"start_time"`      // 开始时间
}

// BatchExtractResult 批量解析结果项
type BatchExtractResult struct {
	FileName        string           `json:"file_name"`
	State           TaskState        `json:"state"`
	ErrorMsg        string           `json:"err_msg,omitempty"`
	FullZipURL      string           `json:"full_zip_url,omitempty"`
	ExtractProgress *ExtractProgress `json:"extract_progress,omitempty"` // 运行中时提供
}

// GetBatchResultResponse 批量获取结果响应
type GetBatchResultResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    struct {
		BatchID       string               `json:"batch_id"`
		ExtractResult []BatchExtractResult `json:"extract_result"`
	} `json:"data"`
	TraceID string `json:"trace_id"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	TraceID string `json:"trace_id"`
}
