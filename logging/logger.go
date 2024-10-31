package logging

import "log"

type Logger interface {
	Log(msg string, data map[string]any)
}

type LoggerFunc func(msg string, data map[string]any)

func (f LoggerFunc) Log(msg string, data map[string]any) {
	f(msg, data)
}
func DefaultLogger(msg string, data map[string]any) {
	log.Println(msg, data)
}
