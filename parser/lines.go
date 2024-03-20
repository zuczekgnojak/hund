package parser

import (
	"bufio"
	"hund/logger"
	"hund/util"
	"os"
	"regexp"
	"strings"
)

const IDENTIFIER = `([a-zA-Z][a-zA-Z0-9_-]*)`
const ARGS_AND_OPTIONS = `((\(.*\))?:.*)`
const OPEN_PARENT_WHITE = `(\([ \t]*\\)`
const ANY_WHITE = `[ \t]*`
const CALL_START = `@\(\(`
const CALL_END = `\)\)`
const EMBED_CALL_START = `@\[\[`
const EMBED_CALL_END = `\]\]`

var targetDefinitionPattern = regexp.MustCompile(`^` + IDENTIFIER + ARGS_AND_OPTIONS)
var escapedNewlineExpression = regexp.MustCompile(`^.*\\$`)
var commentExpression = regexp.MustCompile(`^[ \t]*//.*$`)
var indentedExpression = regexp.MustCompile(`^((  )|\t).*$`)

var headerTargetName = regexp.MustCompile(`^` + IDENTIFIER)
var headerOptionDefinition = regexp.MustCompile(`^` + IDENTIFIER + `(\|[a-zA-Z0-9])?=(value|flag)`)
var headerArgumentDefinition = regexp.MustCompile(`^` + IDENTIFIER + `[\+\?\*]?`)

var variableNameExtractor = regexp.MustCompile(`@{{( )*(?P<name>` + IDENTIFIER + `)( )*}}`)
var callOnlyExpression = regexp.MustCompile(
	`^` + ANY_WHITE + CALL_START + ANY_WHITE + `([a-zA-Z].*)` + CALL_END + ANY_WHITE + `$`,
)
var targetCallExtractor = regexp.MustCompile(CALL_START + ANY_WHITE + `(?P<name>[a-zA-Z].*?)` + ANY_WHITE + CALL_END)
var targetEmbedCallExtractor = regexp.MustCompile(EMBED_CALL_START + ANY_WHITE + `(?P<name>[a-zA-Z].*?)` + ANY_WHITE + EMBED_CALL_END)

var globalNameExtractor = regexp.MustCompile(`^@(?P<name>` + IDENTIFIER + `).*`)
var globalArgsExtractor = regexp.MustCompile(`\((?P<args>.*)\)`)

type Line struct {
	text string
	num  int
}

func (self Line) IsTargetHeader() bool {
	return self.matches(targetDefinitionPattern)
}

func (self Line) IsIndented() bool {
	return self.matches(indentedExpression)
}

func (self Line) IsNewlineEscaped() bool {
	return self.matches(escapedNewlineExpression)
}

func (self Line) IsComment() bool {
	return self.matches(commentExpression)
}

func (self Line) IsEmpty() bool {
	return strings.TrimSpace(self.text) == ""
}

func (self Line) IsGlobal() bool {
	return strings.HasPrefix(self.text, "@")
}

func (self Line) GetGlobalName() (string, error) {
	match := globalNameExtractor.FindStringSubmatch(self.text)
	if match == nil {
		return "", util.NewError("line %d: can't extract global name", self.num)
	}

	return match[globalNameExtractor.SubexpIndex("name")], nil
}

func (self Line) GetGlobalArgs() string {
	match := globalArgsExtractor.FindStringSubmatch(self.text)
	if match == nil {
		return ""
	}
	return match[globalArgsExtractor.SubexpIndex("args")]
}

func (self Line) GetIndentation() (string, error) {
	var spaces, tabs int
	for _, ch := range self.text {
		if ch == ' ' {
			spaces += 1
		} else if ch == '\t' {
			tabs += 1
		} else {
			break
		}
	}

	if spaces > 0 && tabs > 0 {
		return "", util.NewError("line %d: inconsistent indentation, detected %d spaces and %d tabs", self.num, spaces, tabs)
	}
	if spaces == 0 && tabs == 0 {
		return "", util.NewError("line %d: can't detect indentation", self.num)
	}
	if spaces > 0 {
		return strings.Repeat(" ", spaces), nil
	}

	return strings.Repeat("\t", tabs), nil
}

func (self Line) HasIndentation(indentation string) bool {
	if self.IsEmpty() {
		return true
	}
	return strings.HasPrefix(self.text, indentation)
}

type DynamicContent struct {
	col  int
	text string
}

func (self Line) GetVariables() []DynamicContent {
	result := []DynamicContent{}

	indexMatches := variableNameExtractor.FindAllStringIndex(self.text, -1)
	if indexMatches == nil {
		return result
	}

	nameMatches := variableNameExtractor.FindAllStringSubmatch(self.text, -1)
	if nameMatches == nil {
		return result
	}

	for i, match := range nameMatches {
		variable := match[variableNameExtractor.SubexpIndex("name")]
		col := indexMatches[i][0]
		result = append(result, DynamicContent{col: col, text: variable})
	}

	return result
}

type RendererRepr struct {
	Name     string
	InScript string
}

func GetVariables(script string) []RendererRepr {
	result := []RendererRepr{}

	matches := variableNameExtractor.FindAllStringSubmatch(script, -1)
	if matches == nil {
		return result
	}
	for _, match := range matches {
		inScript := match[0]
		name := match[variableNameExtractor.SubexpIndex("name")]
		result = append(result, RendererRepr{Name: name, InScript: inScript})
	}
	return result
}

func GetCalls(script string) []RendererRepr {
	result := []RendererRepr{}

	matches := targetCallExtractor.FindAllStringSubmatch(script, -1)
	if matches == nil {
		return result
	}
	for _, match := range matches {
		inScript := match[0]
		name := match[targetCallExtractor.SubexpIndex("name")]
		result = append(result, RendererRepr{Name: name, InScript: inScript})
	}
	return result
}

func GetEmbeds(script string) []RendererRepr {
	result := []RendererRepr{}

	matches := targetEmbedCallExtractor.FindAllStringSubmatch(script, -1)
	if matches == nil {
		return result
	}
	for _, match := range matches {
		inScript := match[0]
		name := match[targetEmbedCallExtractor.SubexpIndex("name")]
		result = append(result, RendererRepr{Name: name, InScript: inScript})
	}
	return result
}

func (self Line) GetCalls() []DynamicContent {
	result := []DynamicContent{}

	indexMatches := targetCallExtractor.FindAllStringIndex(self.text, -1)
	if indexMatches == nil {
		return result
	}

	nameMatches := targetCallExtractor.FindAllStringSubmatch(self.text, -1)
	if nameMatches == nil {
		return result
	}

	for i, match := range nameMatches {
		variable := match[targetCallExtractor.SubexpIndex("name")]
		col := indexMatches[i][0]
		result = append(result, DynamicContent{col: col, text: variable})
	}

	return result
}

func (self Line) GetEmbedCalls() []DynamicContent {
	result := []DynamicContent{}

	indexMatches := targetEmbedCallExtractor.FindAllStringIndex(self.text, -1)
	if indexMatches == nil {
		return result
	}

	nameMatches := targetEmbedCallExtractor.FindAllStringSubmatch(self.text, -1)
	if nameMatches == nil {
		return result
	}

	for i, match := range nameMatches {
		variable := match[targetEmbedCallExtractor.SubexpIndex("name")]
		col := indexMatches[i][0]
		result = append(result, DynamicContent{col: col, text: variable})
	}

	return result
}

func (self Line) IsCallOnly() bool {
	return self.matches(callOnlyExpression)
}

func (self Line) matches(expression *regexp.Regexp) bool {
	return expression.MatchString(self.text)
}

func ReadFile(filename string) ([]Line, error) {
	result := []Line{}

	logger.Debugf("opening \"%s\" file\n", filename)
	f, err := os.Open(filename)
	if err != nil {
		return result, err
	}
	defer f.Close()

	num := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		num += 1
		text := scanner.Text()
		line := Line{text: text, num: num}
		result = append(result, line)
	}
	if err := scanner.Err(); err != nil {
		return result, err
	}
	logger.Debugf("read %d lines\n", len(result))

	return result, nil
}

type EditableLine struct {
	text string
	col  int
}

func NewEditableLine(s string) *EditableLine {
	return &EditableLine{
		text: s,
		col:  1,
	}
}

func (self *EditableLine) SkipSpaces() {
	for strings.HasPrefix(self.text, " ") {
		self.text = strings.TrimPrefix(self.text, " ")
		self.col += 1
	}
}

func (self *EditableLine) Trim(prefix string) bool {
	if !strings.HasPrefix(self.text, prefix) {
		return false
	}
	self.text = strings.TrimPrefix(self.text, prefix)
	self.col += len(prefix)
	return true
}

func (self *EditableLine) Extract(expression *regexp.Regexp) string {
	value := expression.FindString(self.text)
	self.Trim(value)
	return value
}

func (self *EditableLine) Finished() bool {
	return len(self.text) == 0
}
