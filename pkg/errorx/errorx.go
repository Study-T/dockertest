package errorx

// Code 业务错误码定义
type Code int

const (
	Success Code = 200

	// 服务器错误（50001-50010）
	DatabaseError       Code = 50001
	DatabaseConnection  Code = 50002
	DatabaseQueryFailed Code = 50003

	// 参数验证错误（40001-40010）
	InvalidParameter Code = 40001
	MissingParameter Code = 40002
	InvalidFormat    Code = 40003
)

// GetMessage 获取错误码对应的错误消息
func GetMessage(code Code) string {
	switch code {
	case Success:
		return "success"
	case DatabaseError:
		return "Database error"
	case DatabaseConnection:
		return "Database connection failed"
	case DatabaseQueryFailed:
		return "Database query failed"
	case InvalidParameter:
		return "Invalid parameter"
	case MissingParameter:
		return "Missing required parameter"
	case InvalidFormat:
		return "Invalid format"
	default:
		return "Unknown error"
	}
}

// NewError 创建业务错误
func NewError(code Code) *CodeError {
	return &CodeError{
		code:    int(code),
		message: GetMessage(code),
	}
}

// NewCodeError 创建带自定义消息的业务错误
func NewCodeError(code Code, message string) *CodeError {
	return &CodeError{
		code:    int(code),
		message: message,
	}
}

// CodeError 业务错误结构
type CodeError struct {
	code    int
	message string
	cause   error
}

func (e *CodeError) Error() string {
	if e.cause != nil {
		return "[" + itoa(e.code) + "] " + e.message + ": " + e.cause.Error()
	}
	return "[" + itoa(e.code) + "] " + e.message
}

func (e *CodeError) GetCode() int       { return e.code }
func (e *CodeError) GetMessage() string { return e.message }
func (e *CodeError) Unwrap() error      { return e.cause }

func Wrap(code Code, cause error) *CodeError {
	return &CodeError{
		code:    int(code),
		message: GetMessage(code),
		cause:   cause,
	}
}

func WrapWithMessage(code Code, message string, cause error) *CodeError {
	return &CodeError{
		code:    int(code),
		message: message,
		cause:   cause,
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
