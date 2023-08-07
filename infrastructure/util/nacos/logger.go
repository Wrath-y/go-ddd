package nacos

import "log"

type logger interface {
	Info(...interface{})
	Error(...interface{})
}

type defualtLogger struct {
}

func (l *defualtLogger) Error(v ...interface{}) {
	log.Println(v...)
}

func (l *defualtLogger) Info(v ...interface{}) {
	log.Println(v...)
}
