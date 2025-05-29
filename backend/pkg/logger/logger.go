package logger

import (
	"log"
	"os"
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// New - создает новый экземпляр логгера с настроенными уровнями логирования
func New() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info - выводит информационное сообщение в лог
func (l *Logger) Info(v ...interface{}) {
	l.infoLogger.Println(v...)
}

// Infof - выводит форматированное информационное сообщение в лог
func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Error - выводит сообщение об ошибке в лог
func (l *Logger) Error(v ...interface{}) {
	l.errorLogger.Println(v...)
}

// Errorf - выводит форматированное сообщение об ошибке в лог
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Debug - выводит отладочное сообщение в лог
func (l *Logger) Debug(v ...interface{}) {
	l.debugLogger.Println(v...)
}

// Debugf - выводит форматированное отладочное сообщение в лог
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}

// Fatal - выводит критическое сообщение в лог и завершает программу
func (l *Logger) Fatal(v ...interface{}) {
	l.errorLogger.Fatal(v...)
}

// Fatalf - выводит форматированное критическое сообщение в лог и завершает программу
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.errorLogger.Fatalf(format, v...)
}
