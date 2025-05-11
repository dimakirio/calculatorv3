package logger

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func NewLogger(level string) *Logger {
	return &Logger{log.New(os.Stdout, "", log.LstdFlags)}
}

func (l *Logger) Info(msg string) {
	l.Println("INFO: " + msg)
}

func (l *Logger) Error(msg string) {
	l.Println("ERROR: " + msg)
}

func (l *Logger) Fatal(msg string) {
	l.Println("FATAL: " + msg)
	os.Exit(1)
}
