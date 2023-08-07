package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

func setLogToFile() {
	writer := &lumberjack.Logger{
		Filename:   "log/" + viper.GetString("app.log.topic") + "/app.log",
		MaxSize:    500, // MB
		MaxBackups: 2,   // 备份
		LocalTime:  true,
	}
	handle = func(c *LogItem) {
		buf := bytes.NewBuffer(nil)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(c)
		b := matchReplace(buf.Bytes())
		writer.Write(b) //nolint
	}
}

func setLogToStdout() {
	handle = func(c *LogItem) {
		buf := bytes.NewBuffer(nil)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(c)
		b := matchReplace(buf.Bytes())
		os.Stdout.Write(b) // nolint
	}
}

func setLogToFormat() {
	var colorNum int8
	handle = func(c *LogItem) {
		buf := bytes.NewBuffer(nil)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "\t")
		_ = enc.Encode(c)
		b := matchReplace(buf.Bytes())
		colorNum = (colorNum + 3) & 7
		fmt.Printf("\x1b[0;%dm%s\x1b[0m", colorNum+30, b)
	}
}
