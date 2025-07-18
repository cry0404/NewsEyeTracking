package models

// APIResponse 通用API响应结构
type APIResponse struct {
	Status string      `json:"status"`          // "success" 或 "error"
	Data   interface{} `json:"data,omitempty"`  // 成功时的数据, 任意的结构
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
	ErrorBadRequest			   = "WRONG REQUEST"
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
// api response 在这里可以理解为父亲，error 和 success 都是继承的内容了
// ErrorResponse 创建错误响应，设计的函数，使用 errorinfo 来获取
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

