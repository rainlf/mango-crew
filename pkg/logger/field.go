package logger

import (
	"time"

	"go.uber.org/zap"
)

// Field 日志字段类型
type Field = zap.Field

// String 字符串字段
func String(key, val string) Field {
	return zap.String(key, val)
}

// Int 整数字段
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Int64 长整数字段
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

// Float64 浮点数字段
func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

// Bool 布尔字段
func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

// Duration 时间字段
func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

// Err 错误字段
func Err(err error) Field {
	return zap.Error(err)
}

// Any 任意类型字段
func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}
