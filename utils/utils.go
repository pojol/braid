package utils

import (
	"fmt"
	"runtime"
	"strconv"
)

func MessageWithLine(format string, v ...interface{}) string {
	_, file, line, _ := runtime.Caller(2)
	return shortFile(file) + ":" + strconv.Itoa(line) + ": " + fmt.Sprintf(format, v...)
}

func shortFile(file string) string {
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	return short
}
