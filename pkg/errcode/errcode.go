package errcode

import (
	"errors"
	"fmt"
)

// Error 业务错误码
type Error struct {
	Code    int
	Message string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func (e *Error) WithError(err error) *Error {
	return &Error{Code: e.Code, Message: e.Message, cause: err}
}

func (e *Error) WithErrorf(format string, args ...any) *Error {
	return e.WithError(fmt.Errorf(format, args...))
}

func FromError(err error) *Error {
	var e *Error
	errors.As(err, &e)
	return e
}

// =================== 通用 (10xxx) ===================

var (
	Success       = New(0, "success")
	InvalidParam  = New(10001, "参数无效")
	Unauthorized  = New(10002, "未认证")
	Forbidden     = New(10003, "无权限")
	TooFrequent   = New(10004, "请求过频")
	NotFound      = New(10005, "资源不存在")
	AlreadyExists = New(10006, "资源已存在")
)

// =================== 系统 (50xxx) ===================

var (
	InternalError = New(50001, "服务内部错误")
	DatabaseError = New(50002, "数据库错误")
	CacheError    = New(50003, "缓存错误")
)

// =================== 房源模块 (20xxx) ===================

var (
	RoomNotFound  = New(20001, "房间不存在")
	RoomDisabled  = New(20002, "房间已禁用")
	RoomInvalidId = New(20003, "无效的房间ID")
)
