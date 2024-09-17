package core

import (
	"archive/zip"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"io"
	"os"
	"path/filepath"
	"time"
)

var _ Logger = (*consoleLogger)(nil)

// Logger 日志需要实现的接口定义
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})

	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	// Sync logger Sync calls to flush buffer
	Sync() error
}

// consoleLogger 命令行日志实现
type consoleLogger struct{}

// Debug 日志
func (consoleLogger) Debug(v ...interface{}) {
	output("Debug", fmt.Sprint(v...))
}

// Info 日志
func (consoleLogger) Info(v ...interface{}) {
	output("Info", fmt.Sprint(v...))
}

// Warn 日志
func (consoleLogger) Warn(v ...interface{}) {
	output("Warning", fmt.Sprint(v...))
}

// Error
func (consoleLogger) Error(v ...interface{}) {
	output("Error", fmt.Sprint(v...))
}

// Debugf Debug Format 日志
func (consoleLogger) Debugf(format string, v ...interface{}) {
	output("Debug", fmt.Sprintf(format, v...))
}

// Infof Info Format 日志
func (consoleLogger) Infof(format string, v ...interface{}) {
	output("Info", fmt.Sprintf(format, v...))
}

// Warnf Warning Format 日志
func (consoleLogger) Warnf(format string, v ...interface{}) {
	output("Warning", fmt.Sprintf(format, v...))
}

// Errorf Error Format 日志
func (consoleLogger) Errorf(format string, v ...interface{}) {
	output("Error", fmt.Sprintf(format, v...))
}

// Sync 控制台 logger 不需要 sync
func (consoleLogger) Sync() error {
	return nil
}

func output(level string, v ...interface{}) {
	pc, file, line, _ := runtime.Caller(3)
	file = filepath.Base(file)
	funcName := strings.TrimPrefix(filepath.Ext(runtime.FuncForPC(pc).Name()), ".")

	logFormat := "[%s] %s %s:%d:%s " + fmt.Sprint(v...) + "\n"
	date := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf(logFormat, level, date, file, line, funcName)
}

// DefaultLogger 默认logger
var DefaultLogger = ConsoleFileLogger{
	console: ConsoleLogger,
	file:    FileLogger,
}
var FileLogger = NewFileLogger("./")
var ConsoleLogger = Logger(new(consoleLogger))

// Debug log.Debug
func Debug(v ...interface{}) {
	DefaultLogger.Debug(v...)
}

// Info log.Info
func Info(v ...interface{}) {
	DefaultLogger.Info(v...)
}

// Warn log.Warn
func Warn(v ...interface{}) {
	DefaultLogger.Warn(v...)
}

// Error log.Error
func Errors(v ...interface{}) {
	DefaultLogger.Error(v...)
}

// Debugf log.Debugf
func Debugf(format string, v ...interface{}) {
	DefaultLogger.Debugf(format, v...)
}

// Infof log.Infof
func Infof(format string, v ...interface{}) {
	DefaultLogger.Infof(format, v...)
}

// Warnf log.Warnf
func Warnf(format string, v ...interface{}) {
	DefaultLogger.Warnf(format, v...)
}

// Errorf log.Errorf
func Errorf(format string, v ...interface{}) {
	DefaultLogger.Errorf(format, v...)
}

// Sync logger Sync calls to flush buffer
func Sync() {
	_ = DefaultLogger.Sync()
}

var _ Logger = (*fileLogger)(nil)

// fileLogger 实现文件日志记录
type fileLogger struct {
	basePath string
	mu       sync.Mutex
}

// NewFileLogger 创建一个新的 FileLogger 实例
func NewFileLogger(basePath string) *fileLogger {
	return &fileLogger{basePath: basePath}
}

// Debug 日志
func (f *fileLogger) Debug(v ...interface{}) {
	f.log("Debug", fmt.Sprint(v...))
}

// Info 日志
func (f *fileLogger) Info(v ...interface{}) {
	f.log("Info", fmt.Sprint(v...))
}

// Warn 日志
func (f *fileLogger) Warn(v ...interface{}) {
	f.log("Warning", fmt.Sprint(v...))
}

// Error 日志
func (f *fileLogger) Error(v ...interface{}) {
	f.log("Error", fmt.Sprint(v...))
}

// Debugf 格式化 Debug 日志
func (f *fileLogger) Debugf(format string, v ...interface{}) {
	f.log("Debug", fmt.Sprintf(format, v...))
}

// Infof 格式化 Info 日志
func (f *fileLogger) Infof(format string, v ...interface{}) {
	f.log("Info", fmt.Sprintf(format, v...))
}

// Warnf 格式化 Warning 日志
func (f *fileLogger) Warnf(format string, v ...interface{}) {
	f.log("Warning", fmt.Sprintf(format, v...))
}

// Errorf 格式化 Error 日志
func (f *fileLogger) Errorf(format string, v ...interface{}) {
	f.log("Error", fmt.Sprintf(format, v...))
}

// Sync 关闭当前日志文件
func (f *fileLogger) Sync() error {
	// No-op in this implementation as logs are written directly
	return nil
}

// log 记录日志到文件
func (f *fileLogger) log(level string, message string) {
	now := time.Now()
	date := now.Format("2006-01-02")
	logFile := filepath.Join(f.basePath, "logs", date+".log")

	// Ensure the log directory exists
	err := os.MkdirAll(filepath.Dir(logFile), os.ModePerm)
	if err != nil {
		fmt.Println("Error creating log directory:", err)
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	pc, fileName, line, _ := runtime.Caller(2)
	funcName := filepath.Ext(runtime.FuncForPC(pc).Name())
	fileName = filepath.Base(fileName)

	logFormat := "[%s] %s %s:%d:%s %s\n"
	dateTime := now.Format("2006-01-02 15:04:05")
	_, err = fmt.Fprintf(file, logFormat, level, dateTime, fileName, line, funcName, message)
	if err != nil {
		fmt.Println("Error writing to log file:", err)
	}
}

// RotateLogs 每月压缩上个月的日志文件
func RotateLogs(basePath string) error {
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	monthDir := filepath.Join(basePath, "logs")
	zipFile := filepath.Join(basePath, "logs", "archive", lastMonth.Format("2006-01")+".zip")

	// Ensure the archive directory exists
	err := os.MkdirAll(filepath.Dir(zipFile), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating archive directory: %w", err)
	}

	// Create a new zip file
	archiveFile, err := os.Create(zipFile)
	if err != nil {
		return fmt.Errorf("error creating zip file: %w", err)
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	// Add files to the zip archive
	files, err := filepath.Glob(filepath.Join(monthDir, lastMonth.Format("2006-01")+".log"))
	if err != nil {
		return fmt.Errorf("error finding log files: %w", err)
	}

	for _, file := range files {
		err := addFileToZip(zipWriter, file)
		if err != nil {
			return fmt.Errorf("error adding file to zip: %w", err)
		}
	}

	return nil
}

// addFileToZip 添加文件到 zip 压缩包
func addFileToZip(zipWriter *zip.Writer, file string) error {
	zipFile, err := zipWriter.Create(filepath.Base(file))
	if err != nil {
		return err
	}

	srcFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(zipFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

type ConsoleFileLogger struct {
	console Logger
	file    *fileLogger
}

func (c *ConsoleFileLogger) Debug(v ...interface{}) {
	output("Debug", fmt.Sprint(v...))
	c.file.log("Debug", fmt.Sprint(v...))
}

func (c ConsoleFileLogger) Info(v ...interface{}) {
	output("Info", fmt.Sprint(v...))
	c.file.log("Info", fmt.Sprint(v...))
}

func (c ConsoleFileLogger) Warn(v ...interface{}) {
	output("Warn", fmt.Sprint(v...))
	c.file.log("Warn", fmt.Sprint(v...))

}

func (c ConsoleFileLogger) Error(v ...interface{}) {
	output("Error", fmt.Sprint(v...))
	c.file.log("Error", fmt.Sprint(v...))
}

func (c ConsoleFileLogger) Debugf(format string, v ...interface{}) {
	output("Debug", fmt.Sprintf(format, v...))
	c.file.log("Debug", fmt.Sprintf(format, v...))
}

func (c ConsoleFileLogger) Infof(format string, v ...interface{}) {
	output("Info", fmt.Sprintf(format, v...))
	c.file.log("Info", fmt.Sprintf(format, v...))
}

func (c ConsoleFileLogger) Warnf(format string, v ...interface{}) {
	output("Warn", fmt.Sprintf(format, v...))
	c.file.log("Warn", fmt.Sprintf(format, v...))
}

func (c ConsoleFileLogger) Errorf(format string, v ...interface{}) {
	output("Error", fmt.Sprintf(format, v...))
	c.file.log("Error", fmt.Sprintf(format, v...))
}

func (c ConsoleFileLogger) Sync() error {
	c.console.Sync()
	return c.file.Sync()
}
