package logtlm

import (
	"fmt"
	"os"
	"path"
	"time"
)

// 定义一个日志的结构体
type logger struct {
	method      uint8    // 写日志的方法	1,为控制台输出,2为文件输出
	Level       LogLevel // 日志级别
	filePath    string   // 日志文件路径
	fileName    string   // 日志文件名
	maxFileSize int64    // 日志文件最大大小
	fileObj     *os.File // 日志文件
	errFileObj  *os.File // 错误日志文件
}

// 初始化一个写文件的日志对象
func Newlogger(method, levelStr, fp, fn string, maxSize int64) (Logger, error) {
	// 判断写日志的方法
	if method != "file" && method != "console" {
		return nil, fmt.Errorf("method is invalid")
	}

	// 判断日志级别
	level, err := pasrseLogLevel(levelStr) // 把字符串解析成日志级别int64
	if err != nil {
		return nil, err
	}

	switch method {
	case "console": // 如果是console,则不需要创建文件,控制台输出
		fmt.Println("如果是console,则不需要创建文件,控制台输出")
		return &logger{
			method:      1,
			Level:       level,
			filePath:    "",
			fileName:    "",
			maxFileSize: 0,
			fileObj:     nil,
			errFileObj:  nil,
		}, nil
	case "file":
		// 创建一个日志对象
		f1 := &logger{
			method:      2,
			Level:       level,
			filePath:    fp,
			fileName:    fn,
			maxFileSize: maxSize,
		}
		err = f1.initFile() // 按照文件路径和文件名创建文件或者打开文件或追加内容
		return f1, err
	default:
		return nil, fmt.Errorf("method is invalid")
	}
}
func (l *logger) initFile() error {
	// 创建文件夹
	err := os.MkdirAll(l.filePath, 0666)
	if err != nil {
		fmt.Printf("创建文件夹失败,err:%v\n", err)
		return err
	}
	// 创建文件
	fullFileName := path.Join(l.filePath, l.fileName) // 拼接文件路径和文件名
	fileObj, err := os.OpenFile(fullFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("open log file failed,err :%v\n", err)
		return err
	}
	errFileName := path.Join(l.filePath, l.fileName[:len(l.fileName)-len(path.Ext(l.fileName))]+"_err"+path.Ext(l.fileName))
	errFileObj, err := os.OpenFile(errFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("open err log file failed,err :%v\n", err)
		return err
	}
	// 日志文件都已经打开了，就不需要再打开了
	l.fileObj = fileObj
	l.errFileObj = errFileObj
	return nil
}

// 日志级别开关,即判断是否记录该等级日志
func (l *logger) enable(level LogLevel) bool { // 判断是否开启
	return level >= l.Level
}

// 切割日志文件
// 判断文件是否超过最大大小,需要切割
func (l *logger) fileSpilt(file *os.File) (*os.File, error) {
	// 需要切割的文件
	nowStr := time.Now().Format("20060102150405000")
	stat, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info failed,err: %v\n", err)
	}

	if stat.Size() >= l.maxFileSize { // 如果文件大于最大大小
		logName := path.Join(l.filePath, stat.Name())
		newLogName := fmt.Sprintf("%s.bak%s.log", logName, nowStr)
		// 1,关闭当前文件
		file.Close()
		// 2, 备份当前文件 rename newLogName是老的文件
		os.Rename(logName, newLogName)
		// 3,打开一个新的日志文件
		fileObj, err := os.OpenFile(logName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("打开一个新的日志文件错误,err :%v\n", err)
		}
		return fileObj, nil
	} else {
		// 如果文件小于最大大小,则不需要切割
		return file, nil // 返回原来的文件
	}
}

// 写入日志到文件
func (l *logger) log(lv LogLevel, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	now := time.Now()
	funcName, fileName, line := getInfo(3)
	if l.enable(lv) {
		if l.method == 1 { // 如果是console
			fmt.Printf("[%s] [%s] [%v : %v : %v] : %v\n", lv.logString(), now.Format("2006-01-02 15:04:05"), fileName, funcName, line, msg)
		} else {
			// 用 Fprintf 写入自定义的日志格式到日志文件
			fileObj, err := l.fileSpilt(l.fileObj)
			if err != nil {
				fmt.Printf("切割文件错误,err:%v\n", err)
			}
			l.fileObj = fileObj
			fmt.Fprintf(l.fileObj, "[%s] [%s] [%v : %v : %v] : %v\n", lv.logString(), now.Format("2006-01-02 15:04:05"), fileName, funcName, line, msg)
			if lv >= ERROR { // 如果是错误级别，就再写入一份到错误日志文件
				errFileObj, err := l.fileSpilt(l.errFileObj)
				if err != nil {
					fmt.Printf("切割文件错误,err:%v\n", err)
				}
				l.errFileObj = errFileObj

				fmt.Fprintf(l.errFileObj, "[%s] [%s] [%v : %v : %v] : %v\n", lv.logString(), now.Format("2006-01-02 15:04:05"), fileName, funcName, line, msg)
			}
		}
	}
}

// 直接调用的方法
func (l *logger) Trace(format string, a ...interface{}) {
	l.log(TRACE, format, a...)
}
func (l *logger) Debug(format string, a ...interface{}) {
	l.log(DEBUG, format, a...)
}
func (l *logger) Info(format string, a ...interface{}) {
	l.log(INFO, format, a...)
}
func (l *logger) Warn(format string, a ...interface{}) {
	l.log(WARN, format, a...)
}
func (l *logger) Error(format string, a ...interface{}) {
	l.log(ERROR, format, a...)
}
func (l *logger) Panic(format string, a ...interface{}) {
	l.log(PANIC, format, a...)
	l.Close()
	panic(fmt.Sprintf(format, a...))
}
func (l *logger) Fatal(format string, a ...interface{}) {
	l.log(FATAL, format, a...)
	l.Close()
	os.Exit(1)
}

// 关闭文件
func (l *logger) Close() {
	l.fileObj.Close()
	l.errFileObj.Close()
}
