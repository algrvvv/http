package logger

import (
	"fmt"
	"os"
)

const (
	SuccessLogType = 0
	InfoLogType    = 1
	ErrorLogType   = 2
	ExitLogType    = 3
)

func Logger(log any, logType int) {
	if logType == SuccessLogType {
		fmt.Printf("[%s] - %v\n", Green("SUCC"), log)
	} else if logType == InfoLogType {
		fmt.Printf("[%s] - %v\n", Blue("INFO"), log)
	} else if logType == ErrorLogType {
		fmt.Printf("[%s] - %v\n", Red("ERR"), log)
	} else if logType == ExitLogType {
		fmt.Printf("[%s] - %v\n", Red("EXIT"), log)
		os.Exit(0)
	}
}
