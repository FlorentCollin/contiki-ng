package utils

// Logging module that outputs colored strings to the standard output
// based on a log level.
// This help especially in the terminal to see errors because the output might
// be quite fast when a lot of nodes are sending packet to the server.

import "fmt"

type Color struct {
	r uint8
	g uint8
	b uint8
}

var YELLOW = Color{255, 244, 79}
var RED = Color{202, 52, 51}
var BLUE = Color{122, 162, 247}
var WHITE = Color{255, 255, 255}

type LogLevel int

const (
	LogLevelInfo = iota
	LogLevelDebug
	LogLevelWarning
	LogLevelError
)

func (c *Color) start() string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", c.r, c.g, c.b)
}
func (c *Color) end() string {
	return "\033[0m"
}

type Logger struct {
	logLevel LogLevel
	color    Color
}

var Log *Logger

func NewLogger(logLevel LogLevel, color Color) *Logger {
	if Log == nil {
		Log = &Logger{logLevel: logLevel, color: color}
	}
	return Log
}

func (l *Logger) Print(v ...interface{}) {
	l.ColorPrint(l.color, v...)
}

func (l *Logger) Println(v ...interface{}) {
	l.ColorPrintln(l.color, v...)
}

func (l *Logger) ColorPrint(color Color, v ...interface{}) {
	l.ColorLevelPrint(LogLevelInfo, color, v...)
}

func (l *Logger) ColorPrintln(color Color, v ...interface{}) {
	l.ColorLevelPrintln(LogLevelInfo, color, v...)
}

func (l *Logger) ColorLevelPrint(logLevel LogLevel, color Color, v ...interface{}) {
	if logLevel < l.logLevel {
		return
	}
	fmt.Print(color.start())
	fmt.Print(v...)
	fmt.Print(color.end())
}

func (l *Logger) ColorLevelPrintln(logLevel LogLevel, color Color, v ...interface{}) {
	if logLevel < l.logLevel {
		return
	}
	fmt.Print(color.start())
	fmt.Print(v...)
	fmt.Println(color.end())
}

func (l *Logger) InfoPrint(v ...interface{}) {
	l.ColorLevelPrint(LogLevelInfo, BLUE, v...)
}

func (l *Logger) DebugPrint(v ...interface{}) {
	l.ColorLevelPrint(LogLevelDebug, Color{0, 0, 0}, v...)
}

func (l *Logger) WarningPrint(v ...interface{}) {
	l.ColorLevelPrint(LogLevelWarning, YELLOW, v...)
}

func (l *Logger) ErrorPrint(v ...interface{}) {
	l.ColorLevelPrint(LogLevelError, RED, v...)
}
func (l *Logger) InfoPrintln(v ...interface{}) {
	l.ColorLevelPrintln(LogLevelInfo, BLUE, v...)
}

func (l *Logger) DebugPrintln(v ...interface{}) {
	l.ColorLevelPrintln(LogLevelDebug, Color{0, 0, 0}, v...)
}

func (l *Logger) WarningPrintln(v ...interface{}) {
	l.ColorLevelPrintln(LogLevelWarning, YELLOW, v...)
}

func (l *Logger) ErrorPrintln(v ...interface{}) {
	l.ColorLevelPrintln(LogLevelError, RED, v...)
}
