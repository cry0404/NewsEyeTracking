package models

// APIResponse 通用API响应结构
type APIResponse struct {
	Status string      `json:"status"`          // "success" 或 "error"
	Data   interface{} `json:"data,omitempty"`  // 成功时的数据
	Error  *ErrorInfo  `json:"error,omitempty"` // 错误时的信息
}

// ErrorInfo 错误信息结构
type ErrorInfo struct {
	Code    string `json:"code"`              // 错误代码
	Message string `json:"message"`           // 错误消息
	Details string `json:"details,omitempty"` // 详细错误信息
}

// 常用错误代码
const (
	ErrorCodeInvalidRequest    = "INVALID_REQUEST"
	ErrorCodeUnauthorized      = "UNAUTHORIZED"
	ErrorCodeForbidden         = "FORBIDDEN"
	ErrorCodeNotFound          = "NOT_FOUND"
	ErrorCodeConflict          = "CONFLICT"
	ErrorCodeInternalError     = "INTERNAL_ERROR"
	ErrorCodeInvalidInviteCode = "INVALID_INVITE_CODE"
	ErrorCodeInviteCodeUsed    = "INVITE_CODE_USED"
	ErrorCodeSessionNotFound   = "SESSION_NOT_FOUND"
	ErrorCodeDataProcessError  = "DATA_PROCESS_ERROR"
	ErrorCodeOSSUploadError    = "OSS_UPLOAD_ERROR"
	ErrorCodeDatabaseError     = "DATABASE_ERROR"
)

// SuccessResponse 创建成功响应
func SuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Status: "success",
		Data:   data,
	}
}

// ErrorResponse 创建错误响应
func ErrorResponse(code, message, details string) *APIResponse {
	return &APIResponse{
		Status: "error",
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
/*
// PaginationInfo 分页信息
type PaginationInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Items      interface{}     `json:"items"`
	Pagination *PaginationInfo `json:"pagination"`
}
*/

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// VersionResponse 版本信息响应
type VersionResponse struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
}
