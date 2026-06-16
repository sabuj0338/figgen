package logger

import (
	"fmt"
	"os"
	"time"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	gray    = "\033[90m"
)

func timestamp() string {
	return gray + "[" + time.Now().Format("15:04:05") + "]" + reset
}

func formatMessage(prefixColor, prefix, msg string) string {
	return fmt.Sprintf("%s %s %s | %s %s", timestamp(), prefixColor+bold, prefix, reset, msg)
}

// Info prints a general informational message
func Info(format string, a ...interface{}) {
	fmt.Println(formatMessage(cyan, "INFO ", fmt.Sprintf(format, a...)))
}

// Success prints a success message
func Success(format string, a ...interface{}) {
	fmt.Println(formatMessage(green, "SUCC ", fmt.Sprintf(format, a...)))
}

// Warn prints a warning message
func Warn(format string, a ...interface{}) {
	fmt.Println(formatMessage(yellow, "WARN ", fmt.Sprintf(format, a...)))
}

// Error prints an error message
func Error(format string, a ...interface{}) {
	fmt.Println(formatMessage(red, "ERROR", fmt.Sprintf(format, a...)))
}

// Step prints a process step
func Step(format string, a ...interface{}) {
	fmt.Println(formatMessage(magenta, "STEP ", fmt.Sprintf(format, a...)))
}

// Debug prints debug logs (can be used for verbose debugging)
func Debug(format string, a ...interface{}) {
	fmt.Println(formatMessage(gray, "DEBUG", fmt.Sprintf(format, a...)))
}

// Fatal prints an error and exits the program
func Fatal(format string, a ...interface{}) {
	Error(format, a...)
	os.Exit(1)
}

// Prompt prints a yellow prompt without a newline
func Prompt(format string, a ...interface{}) {
	fmt.Print(formatMessage(yellow, "INPUT", fmt.Sprintf(format, a...)))
}
