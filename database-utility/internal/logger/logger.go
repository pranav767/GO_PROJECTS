package logger

import (
	"log"
	"os"
)

var (
	defaultLogger = log.New(os.Stdout, "", log.LstdFlags)
)

func Info(msg string)  { defaultLogger.Printf("INFO: %s", msg) }
func Error(msg string) { defaultLogger.Printf("ERROR: %s", msg) }

func SetOutputFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defaultLogger = log.New(f, "", log.LstdFlags)
	return nil
}
