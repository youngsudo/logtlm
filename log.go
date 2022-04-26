package logtlm

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// 定义一个接口
type Logger interface {
	Debug(format string, a ...interface{})
	Trace(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Error(format string, a ...interface{})
	Panic(format string, a ...interface{})
	Fatal(format string, a ...interface{})
}

type LogLevel uint16 // 日志级别

const (
	UNKNOWN LogLevel = iota // 未知级别
	DEBUG
	TRACE
	INFO
	WARN
	ERROR
	PANIC
	FATAL
)

// 解析日志级别 字符串转换为数字 如果转换失败返回错误
func pasrseLogLevel(level string) (LogLevel, error) {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		return DEBUG, nil
	case "trace":
		return TRACE, nil
	case "info":
		return INFO, nil
	case "warn":
		return WARN, nil
	case "error":
		return ERROR, nil
	case "panic":
		return PANIC, nil
	case "fatal":
		return FATAL, nil
	default:
		return UNKNOWN, errors.New("日志级别不存在")
	}
}

// 将日志级别数字转string
func (l LogLevel) logString() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// getInfo  获取调用者信息 skip 表示调用栈的层数
func getInfo(skip int) (funcName, fileName string, line int) {
	pc, file, line, ok := runtime.Caller(skip) // 获取调用者的信息
	if !ok {
		fmt.Println("runtime.Caller() failed")
		return
	}
	funcName = runtime.FuncForPC(pc).Name()
	funcName = strings.Split(funcName, ".")[1]       // 获取函数名
	fileName = file[strings.LastIndex(file, "/")+1:] // 文件名
	return
}
