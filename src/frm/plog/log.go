/*
 * 自己实现log库而不使用标准log库的原因是我们对log有一些自己的定制化需求,比如：
 * 1. log分级
 * 2. 每日log需写入到一个文件，每月log需要归拢到一个文件夹
 * 3. 根据运维要求进行小幅修改
 * 系统的log太过简单。而第三方库引入未知复杂度且不一定能解决问题
 * 因此，根据需求实现一个简单版本,底层仍然使用标准log库
 */
package plog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/fatih/color"
)

var logLevel LogLevel = LOG_INFO

type LogLevel int

const (
	LOG_TRACE LogLevel = 0
	LOG_DEBUG LogLevel = 1
	LOG_INFO  LogLevel = 2
	LOG_ERROR LogLevel = 3
	LOG_CRIT  LogLevel = 4
	LOG_PANIC LogLevel = 5
	LOG_FATAL LogLevel = 6
)

func InitLog(logFile string, level LogLevel) {
	logLevel = level
	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		Fatal("Can't open log file")
	}
	log.SetOutput(file)
}

func Trace(v ...interface{}) {
	if logLevel > LOG_TRACE {
		return
	}

	var prefix string
	timenow := time.Now().Format("15:04:05")
	if funcPtr, filePath, line, ok := runtime.Caller(1); ok {
		prefix = fmt.Sprintf("%s:%d %s -> [Trace]", filepath.Base(filePath), line, runtime.FuncForPC(funcPtr).Name())
	}
	color.Set(color.FgBlue)
	fmt.Printf("%s %s ", timenow, prefix)
	fmt.Println(v...)
	color.Unset()
	log.Println(append([]interface{}{timenow, prefix}, v...)...)
}

func Tracef(format string, v ...interface{}) {
	if logLevel > LOG_TRACE {
		return
	}

	var prefix string
	timenow := time.Now().Format("15:04:05")
	if funcPtr, filePath, line, ok := runtime.Caller(1); ok {
		prefix = fmt.Sprintf("%s:%d %s -> [Trace]", filepath.Base(filePath), line, runtime.FuncForPC(funcPtr).Name())
	}
	color.Set(color.FgBlue)
	fmt.Printf("%s %s ", timenow, prefix)
	fmt.Printf(format, v...)
	color.Unset()
	log.Println(append([]interface{}{timenow, prefix}, v...)...)
}

func Debug(v ...interface{}) {
	if logLevel > LOG_DEBUG {
		return
	}

	var prefix string
	timenow := time.Now().Format("15:04:05")
	if funcPtr, filePath, line, ok := runtime.Caller(1); ok {
		prefix = fmt.Sprintf("%s:%d %s -> [DEBUG]", filepath.Base(filePath), line, runtime.FuncForPC(funcPtr).Name())
	}
	color.Set(color.FgCyan)
	fmt.Printf("%s %s ", timenow, prefix)
	fmt.Println(v...)
	color.Unset()
	log.Println(append([]interface{}{timenow, prefix}, v...)...)
}

func Debugf(format string, v ...interface{}) {
	if logLevel > LOG_DEBUG {
		return
	}

	var prefix string
	timenow := time.Now().Format("15:04:05")
	if funcPtr, filePath, line, ok := runtime.Caller(1); ok {
		prefix = fmt.Sprintf("%s:%d %s -> [DEBUG]", filepath.Base(filePath), line, runtime.FuncForPC(funcPtr).Name())
	}
	color.Set(color.FgCyan)
	fmt.Printf("%s %s ", timenow, prefix)
	fmt.Printf(format, v...)
	color.Unset()
	log.Printf("%s %s"+format, append([]interface{}{timenow, prefix}, v...)...)
}

func Info(v ...interface{}) {
	if logLevel > LOG_INFO {
		return
	}

	var prefix string
	timenow := time.Now().Format("15:04:05")
	if funcPtr, filePath, line, ok := runtime.Caller(1); ok {
		prefix = fmt.Sprintf("%s:%d %s -> [INFO]", filepath.Base(filePath), line, runtime.FuncForPC(funcPtr).Name())
	}
	color.Set(color.FgYellow)
	fmt.Printf("%s %s ", timenow, prefix)
	fmt.Println(v...)
	color.Unset()
	log.Println(append([]interface{}{timenow, prefix}, v...)...)
}

func Infof(format string, v ...interface{}) {
	if logLevel > LOG_INFO {
		return
	}

	var prefix string
	timenow := time.Now().Format("15:04:05")
	if funcPtr, filePath, line, ok := runtime.Caller(1); ok {
		prefix = fmt.Sprintf("%s:%d %s -> [INFO]", filepath.Base(filePath), line, runtime.FuncForPC(funcPtr).Name())
	}
	color.Set(color.FgYellow)
	fmt.Printf("%s %s ", timenow, prefix)
	fmt.Printf(format, v...)
	color.Unset()
	log.Printf("%s %s"+format, append([]interface{}{timenow, prefix}, v...)...)
}

func Error(v ...interface{}) {
	if logLevel > LOG_ERROR {
		return
	}

	var prefix string
	timenow := time.Now().Format("15:04:05")
	if funcPtr, filePath, line, ok := runtime.Caller(1); ok {
		prefix = fmt.Sprintf("%s:%d %s -> [ERROR]", filepath.Base(filePath), line, runtime.FuncForPC(funcPtr).Name())
	}
	color.Set(color.FgRed)
	fmt.Printf("%s %s ", timenow, prefix)
	fmt.Println(v...)
	color.Unset()
	log.Println(append([]interface{}{timenow, prefix}, v...)...)
}

func Errorf(format string, v ...interface{}) {
	if logLevel > LOG_ERROR {
		return
	}

	var prefix string
	timenow := time.Now().Format("15:04:05")
	if funcPtr, filePath, line, ok := runtime.Caller(1); ok {
		prefix = fmt.Sprintf("%s:%d %s -> [ERROR]", filepath.Base(filePath), line, runtime.FuncForPC(funcPtr).Name())
	}
	color.Set(color.FgRed)
	fmt.Printf("%s %s ", timenow, prefix)
	fmt.Printf(format, v...)
	color.Unset()
	log.Printf("%s %s"+format, append([]interface{}{timenow, prefix}, v...)...)
}

func Critical(v ...interface{}) {
	if logLevel > LOG_CRIT {
		return
	}

	var prefix string
	timenow := time.Now().Format("15:04:05")
	if funcPtr, filePath, line, ok := runtime.Caller(1); ok {
		prefix = fmt.Sprintf("%s:%d %s -> [CRIT]", filepath.Base(filePath), line, runtime.FuncForPC(funcPtr).Name())
	}
	color.Set(color.FgRed)
	fmt.Println(append([]interface{}{timenow, prefix}, v...)...)
	color.Unset()
	log.Println(append([]interface{}{timenow, prefix}, v...)...)
}

func Fatal(v ...interface{}) {
	if logLevel > LOG_FATAL {
		return
	}

	if funcPtr, file, line, ok := runtime.Caller(1); ok {
		log.Print("  "+runtime.FuncForPC(funcPtr).Name(), " ", file, " ", line, "\n")
	}
	log.Fatal(v...)
}

func Panic(v ...interface{}) {
	if logLevel > LOG_PANIC {
		return
	}

	log.Panic(v...)
}
