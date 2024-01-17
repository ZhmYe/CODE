package Logger

import (
	"bufio"
	"fmt"
	"os"
)

type Logger struct {
	path   string
	writer *bufio.Writer
	file   *os.File
}

func NewLogger(path string) *Logger {
	logger := new(Logger)
	logger.path = path
	var err error
	logger.file, err = os.OpenFile(logger.path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("open file err=%v\n", err)
		panic(err)
	}
	writer := bufio.NewWriter(logger.file)
	logger.writer = writer
	return logger
}
func (l *Logger) Write(str string) {
	_, err := l.writer.WriteString(str)
	if err != nil {
		return
	}
}
func (l *Logger) Finish() {
	err := l.writer.Flush()
	if err != nil {
		return
	}
	err = l.file.Close()
	if err != nil {
		return
	}
}
