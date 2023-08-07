package logging

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type level int8

const (
	_ level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
)

const timeFormat = "2006-01-02 15:04:05.000000"

var (
	once   = &sync.Once{}
	handle = func(item *LogItem) {}
	sep    = regexp.MustCompile(`\\*"`)
	secret []*regexp.Regexp
	block  cipher.Block
)

type LoggerI interface {
	SetRequestID(requestID string)
	Setv1(v1 string)
	Setv2(v2 string)
	Setv3(v3 string)
	Info(message string, request any, response any, opts ...AttrOption)
	Warn(message string, request any, response any, opts ...AttrOption)
	ErrorL(message string, request any, response any, opts ...AttrOption)
	Fatal(message string, request any, response any, opts ...AttrOption)
}

type Logger struct {
	RequestID string
	V1        string
	V2        string
	V3        string
}

type LogItem struct {
	Level       level  `json:"level"`
	RequestID   string `json:"request_id"`
	V1          string `json:"v1"`
	V2          string `json:"v2"`
	V3          string `json:"v3"`
	Message     string `json:"message"`
	Request     any    `json:"request"`
	Response    any    `json:"response"`
	CreateTime  string `json:"create_time"`
	ExecuteTime int64  `json:"execute_time"`
}

type AttrOption struct {
	RequestID string     `json:"request_id"`
	StartTime *time.Time `json:"start_time"`
}

func Setup() {
	once.Do(func() {
		fields := viper.GetStringSlice("app.aes.fields")
		aesKey := viper.GetString("app.aes.key")
		if len(fields) > 0 && aesKey != "" {
			b, err := aes.NewCipher([]byte(aesKey))
			if err != nil {
				log.Fatal(err)
			}
			block = b
			valid := regexp.MustCompile(`^\w+$`) //字段名仅可包含字母数字下划线
			for _, f := range fields {
				if !valid.MatchString(f) {
					log.Fatal("invalid log.cipher.field: ", f)
				}
				secret = append(secret, regexp.MustCompile(`(?i)\\*"\w*`+f+`\w*\\*"\s*:\s*\\*"(.*?)\\*"`))
			}
		}

		switch viper.GetString("app.log.output") {
		case "file":
			setLogToFile() //每行一条紧凑的无格式json输出到文件
		case "fmt":
			setLogToFormat() //每条日志以带缩进的格式化json输出到控制台，相邻日志颜色不同
		default:
			setLogToStdout() //每行一条紧凑的无格式json输出到控制台
		}
	})
}

func New() *Logger {
	return &Logger{}
}

func NewV(v1, v2, v3 string) *Logger {
	return &Logger{
		V1: v1,
		V2: v2,
		V3: v3,
	}
}

func NewLogger(reqid, v1, v2, v3 string) *Logger {
	return &Logger{
		RequestID: reqid,
		V1:        v1,
		V2:        v2,
		V3:        v3,
	}
}

func (l *Logger) SetRequestID(requestID string) {
	l.RequestID = requestID
}

func (l *Logger) Setv1(v1 string) {
	l.V1 = v1
}

func (l *Logger) Setv2(v2 string) {
	l.V2 = v2
}

func (l *Logger) Setv3(v3 string) {
	l.V3 = v3
}

func convert(val any) any {
	switch v := val.(type) {
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	case []byte:
		return string(v)
	default:
		return v
	}
}

func (l *Logger) stash(lev level, msg string, req, resp any, opts ...AttrOption) {
	logs := &LogItem{
		Level:      lev,
		RequestID:  l.RequestID,
		V1:         l.V1,
		V2:         l.V2,
		V3:         l.V3,
		Message:    msg,
		Request:    convert(req),
		Response:   convert(resp),
		CreateTime: time.Now().Format(timeFormat),
	}
	if logs.V2 == "" {
		if file, line, ok := getFilterCallers(); ok {
			logs.V2 = file + ":" + strconv.Itoa(line)
		}
	}
	if len(opts) > 0 {
		opt := opts[0]
		if logs.RequestID == "" {
			logs.RequestID = opt.RequestID
		}
		if opt.StartTime != nil {
			logs.ExecuteTime = time.Since(*opt.StartTime).Milliseconds()
		}
	}
	handle(logs)
}

func (l *Logger) Info(message string, request, response any, opts ...AttrOption) {
	l.stash(InfoLevel, message, request, response, opts...)
}

func (l *Logger) Warn(message string, request, response any, opts ...AttrOption) {
	l.stash(WarnLevel, message, request, response, opts...)
}

func (l *Logger) ErrorL(message string, request, response any, opts ...AttrOption) {
	l.stash(ErrorLevel, message, request, response, opts...)
}

func (l *Logger) Fatal(message string, request, response any, opts ...AttrOption) {
	l.stash(FatalLevel, message, request, response, opts...)
}

func Info(reqID, v1, v2, v3, message string, request, response any, opts ...AttrOption) {
	NewLogger(reqID, v1, v2, v3).stash(InfoLevel, message, request, response, opts...)
}

func Warn(reqID, v1, v2, v3, message string, request, response any, opts ...AttrOption) {
	NewLogger(reqID, v1, v2, v3).stash(WarnLevel, message, request, response, opts...)
}

func Error(reqID, v1, v2, v3, message string, request, response any, opts ...AttrOption) {
	NewLogger(reqID, v1, v2, v3).stash(ErrorLevel, message, request, response, opts...)
}

func Fatal(reqID, v1, v2, v3, message string, request, response any, opts ...AttrOption) {
	NewLogger(reqID, v1, v2, v3).stash(FatalLevel, message, request, response, opts...)
}

// getFilterCallers 获取过滤后的调用栈
func getFilterCallers() (file string, line int, ok bool) {
	for i := 2; i < 6; i++ {
		_, file, line, ok = runtime.Caller(i)

		if !ok {
			continue
		}
		if strings.Index(file, "core/handle.go") > 0 || strings.Index(file, "logging/logging.go") > 0 {
			continue
		} else {
			return file, line, ok
		}
	}
	return
}
