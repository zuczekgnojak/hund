package logger

import (
	"fmt"
	"hund/util"
	"log"
	"os"
)

var logger *log.Logger = nil
var errLogger *log.Logger = log.New(os.Stderr, "", 0)

func SetVerbose(verbose bool) {
	if !verbose {
		logger = nil
		return
	}
	logger = log.New(os.Stdout, "", 0)
}

func Debugf(format string, v ...any) {
	if logger == nil {
		return
	}

	msg := fmt.Sprintf(format, v...)
	prefix := getDebugPrefix()
	logger.Printf("%s %s", prefix, msg)
}

func Debugln(v ...any) {
	if logger == nil {
		return
	}

	msg := fmt.Sprintln(v...)
	prefix := getDebugPrefix()
	logger.Printf("%s %s\n", prefix, msg)

}

func getDebugPrefix() string {
	callerInfo := util.GetCallerInfo(2)
	return fmt.Sprintf("[DEBUG][%s]", callerInfo)
}

func Errorf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	prefix := getErrorPrefix()
	logger.Printf("%s %s", prefix, msg)
}

func Errorln(v ...any) {
	msg := fmt.Sprintln(v...)
	prefix := getErrorPrefix()
	logger.Printf("%s %s\n", prefix, msg)

}

func Error(err error) {
	prefix := getErrorPrefix()
	errLogger.Printf("%s %s\n", prefix, err.Error())
}

func getErrorPrefix() string {
	callerInfo := util.GetCallerInfo(2)
	return fmt.Sprintf("[ERROR][%s]", callerInfo)
}
