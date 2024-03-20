package util

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

func Quote(s string) string {
	return "\"" + s + "\""
}

func EscapeNL(s string) string {
	return strings.ReplaceAll(s, "\n", "\\n")
}

func StringToArgs(s string) []string {
	result := []string{}
	// https://stackoverflow.com/questions/171480/regex-grabbing-values-between-quotation-marks#comment66257229_171499
	quotedRegex := regexp.MustCompile(`^"(?:\\.|[^\\])*?"`)

	for {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			break
		}

		match := quotedRegex.FindString(s)

		if match == "" {
			word, rest, _ := strings.Cut(s, " ")
			result = append(result, word)
			s = rest
			continue
		}

		s = strings.TrimPrefix(s, match)
		match = strings.Trim(match, "\"")
		result = append(result, match)
	}

	return result
}

func GetCallerInfo(level int) string {
	_, file, line, _ := runtime.Caller(level + 1)
	pathElements := strings.Split(file, "/")
	filename := pathElements[len(pathElements)-1]
	return fmt.Sprintf("%s:%d", filename, line)
}

func NewError(format string, v ...any) error {
	format = format + " (%s)"
	callerInfo := GetCallerInfo(1)
	v = append(v, callerInfo)
	return fmt.Errorf(format, v...)
}
