package parser

import (
	"hund/cli"
	"hund/hundfile"
	"hund/logger"
	"hund/util"
	"strings"
)

type HundfileParser struct {
	currentLine int
	linesNum    int
	globalLines []Line
	lines       []Line
	targets     []*TargetParseStruct
	hundfile    hundfile.Hundfile
}

type ParseFunc func(Line) (bool, error)
type ValidateFunc func(int) error

type TargetParseStruct struct {
	name     string
	parser   *cli.CliParser
	startNum int
	header   Line
	body     []Line
}

func NewTargetParseStruct(header Line, body []Line) (TargetParseStruct, error) {
	result := TargetParseStruct{}

	result.startNum = header.num
	result.header = header
	result.body = body
	result.parser = cli.NewCliParser()
	return result, nil
}

func NewHundfileParser() *HundfileParser {

	parser := HundfileParser{
		currentLine: 0,
		linesNum:    0,
		lines:       nil,
	}

	logger.Debugf("created parser")
	return &parser
}

func (self *HundfileParser) Parse(lines []Line) (hundfile.Hundfile, error) {
	logger.Debugf("parsing started")
	self.lines = lines
	self.linesNum = len(lines)
	self.targets = []*TargetParseStruct{}
	self.hundfile = hundfile.NewHundfile()

	if self.linesNum == 0 {
		return self.hundfile, util.NewError("cannot parse empty data")
	}

	err := self.validate([]ValidateFunc{
		self.extractGlobals,
		self.applyGlobals,
		self.splitTargets,
		self.parseHeaders,
		self.clearEmptyPreAndPost,
		self.checkEmptyBodies,
		self.checkIndentation,
		self.checkVariables,
		self.checkCalls,
		self.checkEmbedCalls,
		self.addTargets,
	})

	return self.hundfile, err
}

func (self *HundfileParser) extractGlobals(phase int) error {
	logger.Debugf("phase %d: extracting global declarations", phase)
	f := func(line Line) (bool, error) {
		if line.IsEmpty() {
			logger.Debugf("line %d: empty, skipping", line.num)
			return true, nil
		}

		if !line.IsGlobal() {
			return false, nil
		}

		self.globalLines = append(self.globalLines, line)
		return true, nil
	}

	return self.apply(f)
}

func (self *HundfileParser) applyGlobals(phase int) error {
	logger.Debugf("phase %d: applying global directives", phase)

	for _, line := range self.globalLines {
		globalName, err := line.GetGlobalName()
		if err != nil {
			return err
		}
		globalArgs := line.GetGlobalArgs()

		logger.Debugf("line %d: extracted global \"%s\" with args \"%s\"", line.num, globalName, globalArgs)

		err = self.hundfile.ApplyGlobal(globalName, globalArgs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *HundfileParser) splitTargets(phase int) error {
	logger.Debugf("phase %d: splitting targets", phase)
	for self.currentLine < len(self.lines) {
		var header Line
		body := []Line{}
		targetHeaderSet := false
		f := func(line Line) (bool, error) {
			if line.IsTargetHeader() {
				if targetHeaderSet {
					logger.Debugf("line %d: new target detected", line.num)
					return false, nil
				}
				logger.Debugf("line %d: target header", line.num)
				header = line
				targetHeaderSet = true
				return true, nil
			}

			if !line.IsEmpty() && !line.IsIndented() {
				return false, util.NewError("line %d: expected indentation", line.num)
			}

			body = append(body, line)
			logger.Debugf("line %d: target body", line.num)
			return true, nil
		}
		err := self.apply(f)
		if err != nil {
			return err
		}
		targetStruct, err := NewTargetParseStruct(header, body)
		if err != nil {
			return err
		}
		self.targets = append(self.targets, &targetStruct)
	}
	return nil
}

func (self *HundfileParser) parseHeaders(phase int) error {
	logger.Debugf("phase %d: parsing target headers", phase)
	for _, targetRepr := range self.targets {
		err := self.parseHeader(targetRepr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *HundfileParser) checkNameUniquness(phase int) error {
	logger.Debugf("phase %d: checking target name uniquness", phase)
	return nil
}

func (self *HundfileParser) clearEmptyPreAndPost(phase int) error {
	logger.Debugf("phase %d: trimming empty lines before and after target code", phase)
	for _, targetRepr := range self.targets {
		var emptyPre, emptyPost int
		body := targetRepr.body
		for _, line := range body {
			if !line.IsEmpty() {
				break
			}
			emptyPre += 1
		}

		logger.Debugf("detected %d empty pre lines for target \"%s\"", emptyPre, targetRepr.name)

		body = body[emptyPre:]
		for i := len(body) - 1; i >= 0; i-- {
			if !body[i].IsEmpty() {
				break
			}
			emptyPost += 1
		}
		logger.Debugf("detected %d empty post lines for target \"%s\"", emptyPost, targetRepr.name)

		nonEmpty := len(body) - emptyPost
		body = body[:nonEmpty]

		targetRepr.body = body
	}
	return nil
}

func (self *HundfileParser) checkEmptyBodies(phase int) error {
	logger.Debugf("phase %d: checking for empty targets", phase)
perTarget:
	for _, targetRepr := range self.targets {
		for _, line := range targetRepr.body {
			if !line.IsEmpty() {
				continue perTarget
			}
		}
		return util.NewError("line %d: empty target", targetRepr.startNum)
	}
	return nil
}

func (self *HundfileParser) checkIndentation(phase int) error {
	logger.Debugf("phase %d: checking indentation consistency", phase)
	for _, targetRepr := range self.targets {
		firstLine := targetRepr.body[0]
		indentation, err := firstLine.GetIndentation()
		logger.Debugf("line %d: \"%s\" detected indentation of \"%s\" len(%d)", firstLine.num, targetRepr.name, indentation, len(indentation))
		if err != nil {
			return err
		}
		for _, line := range targetRepr.body {
			if !line.HasIndentation(indentation) {
				return util.NewError("line %d: inconsistent indentation", line.num)
			}
		}

	}
	return nil
}

func (self *HundfileParser) checkVariables(phase int) error {
	logger.Debugf("phase %d: checking proper variables usage", phase)
	for _, target := range self.targets {
		logger.Debugf("target \"%s\": checking variables", target.name)
		body := target.body
		parser := target.parser
		targetVariables := make(map[string]bool)
		for _, line := range body {
			variables := line.GetVariables()
			if len(variables) == 0 {
				logger.Debugf("line %d: no variables found", line.num)
				continue
			}
			logger.Debugf("line %d: found %d variables", line.num, len(variables))

			for _, v := range variables {
				targetVariables[v.text] = true
				if !parser.Contains(v.text) {
					return util.NewError("line %d, col %d: undefined variable \"%s\"", line.num, v.col, v.text)
				}
			}
		}

		parserVariables := parser.Names()
		for _, parserVar := range parserVariables {
			if !targetVariables[parserVar] {
				return util.NewError("line %d: target \"%s\" defines unused variable \"%s\"", target.startNum, target.name, parserVar)
			}
		}
	}
	return nil
}

func (self *HundfileParser) checkCalls(phase int) error {
	logger.Debugf("phase %d: checking target calls", phase)
	for _, target := range self.targets {
		logger.Debugf("target \"%s\": checking calls", target.name)
		body := target.body
		for _, line := range body {
			calls := line.GetCalls()
			if len(calls) == 0 {
				logger.Debugf("line %d: no calls found", line.num)
				continue
			}
			logger.Debugf("line %d: found %d calls", line.num, len(calls))
			if len(calls) > 1 {
				return util.NewError("line %d: multiple non-embed calls", line.num)
			}
			call := calls[0]
			if !line.IsCallOnly() {
				return util.NewError("line %d, col %d: non-embed call in embed context", line.num, call.col)
			}

			args := util.StringToArgs(call.text)
			logger.Debugf("line %d: detected call %v", line.num, args)

			if len(args) == 0 {
				return util.NewError("line %d, col %d: detected call but cannot parse arguments", line.num, call.col)
			}

			targetName := args[0]
			args = args[1:]

			logger.Debugf("line %d: looking for target \"%s\"", line.num, targetName)

			var foundTarget *TargetParseStruct
			for _, target := range self.targets {
				if target.name == targetName {
					foundTarget = target
				}
			}

			if foundTarget == nil {
				return util.NewError("line %d, col %d: couldn't find target \"%s\"", line.num, call.col, targetName)
			}
			args, err := foundTarget.parser.Parse(args, cli.NewDummyWriter())
			if err != nil {
				return util.NewError("line %d: invalid target call arguments %w", line.num, err)
			}
			logger.Debugf("args left %v", args)
			if len(args) != 0 {
				return util.NewError("line %d: invalid target call arguments, args left %v", line.num, args)
			}
		}
	}
	return nil
}

func (self *HundfileParser) checkEmbedCalls(phase int) error {
	logger.Debugf("phase %d: checking target embed calls", phase)
	for _, target := range self.targets {
		logger.Debugf("target \"%s\": checking embed calls", target.name)
		body := target.body
		for _, line := range body {
			calls := line.GetEmbedCalls()
			if len(calls) == 0 {
				logger.Debugf("line %d: no embed calls found", line.num)
				continue
			}
			logger.Debugf("line %d: found %d calls", line.num, len(calls))
			for _, call := range calls {
				args := util.StringToArgs(call.text)
				logger.Debugf("line %d: detected call %v", line.num, args)

				if len(args) == 0 {
					return util.NewError("line %d, col %d: detected call but cannot parse arguments", line.num, call.col)
				}

				targetName := args[0]
				args = args[1:]
				logger.Debugf("line %d: looking for target \"%s\"", line.num, targetName)

				var foundTarget *TargetParseStruct
				for _, target := range self.targets {
					if target.name == targetName {
						foundTarget = target
					}
				}

				if foundTarget == nil {
					return util.NewError("line %d, col %d: couldn't find target \"%s\"", line.num, call.col, targetName)
				}
				args, err := foundTarget.parser.Parse(args, cli.NewDummyWriter())
				if err != nil {
					return util.NewError("line %d: invalid target call arguments %w", line.num, err)
				}
				logger.Debugf("args left %v", args)
				if len(args) != 0 {
					return util.NewError("line %d: invalid target call arguments, args left %v", line.num, args)
				}
			}
		}
	}
	return nil
}

func (self *HundfileParser) addTargets(phase int) error {
	logger.Debugf("phase %d: adding targets to hundfile", phase)
	for _, targetSpec := range self.targets {
		target := hundfile.Target{}
		target.Name = targetSpec.name
		target.Parser = targetSpec.parser

		script := []string{}

		firstLine := targetSpec.body[0]
		indentation, err := firstLine.GetIndentation()
		if err != nil {
			return err
		}
		for _, line := range targetSpec.body {
			text := line.text
			text = strings.TrimPrefix(text, indentation)
			script = append(script, text)
		}
		target.Script = strings.Join(script, "\n")

		self.hundfile.AddTarget(target)
	}
	return nil
}

func (self *HundfileParser) parseHeader(targetRepr *TargetParseStruct) error {
	header := targetRepr.header
	logger.Debugf("line %d: parsing header \"%s\"", header.num, header.text)
	line := NewEditableLine(header.text)

	name := line.Extract(headerTargetName)
	if name == "" {
		return util.NewError("line %d, col %d: can't extract target name", header.num, line.col)
	}

	logger.Debugf("line %d: extracted target name \"%s\"", header.num, name)
	targetRepr.name = name

	ok := line.Trim("(")
	if ok {
		logger.Debugf("line %d: detected arguments list", header.num)

		expectNext := false
		for {
			line.SkipSpaces()
			argument := line.Extract(headerArgumentDefinition)
			if argument == "" {
				if expectNext {
					return util.NewError("line %d, col %d: expected argument definition", header.num, line.col)
				}
				break
			}
			err := targetRepr.parser.Add(argument)
			if err != nil {
				return err
			}
			line.SkipSpaces()
			expectNext = line.Trim(",")
		}

		ok = line.Trim(")")
		if !ok {
			return util.NewError("line %d, col %d: expected ')'", header.num, line.col)
		}
	}

	ok = line.Trim(":")
	if !ok {
		return util.NewError("line %d, col %d: expected ':'", header.num, line.col)
	}

	for !line.Finished() {
		line.SkipSpaces()
		option := line.Extract(headerOptionDefinition)
		if option == "" {
			return util.NewError("line %d, col %d: expected option definition", header.num, line.col)
		}

		err := targetRepr.parser.Add(option)
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *HundfileParser) apply(parseFunc ParseFunc) error {
	for self.currentLine < len(self.lines) {
		line := self.lines[self.currentLine]
		if line.IsComment() {
			logger.Debugf("line %d: comment, removed", line.num)
			self.currentLine += 1
			continue
		}
		ok, err := parseFunc(line)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		self.currentLine += 1
	}
	return nil
}

func (self *HundfileParser) validate(funcs []ValidateFunc) error {
	for i, f := range funcs {
		err := f(i + 1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *HundfileParser) getCurrentLine() Line {
	return self.lines[self.currentLine]
}

func skipEmptyAndComments(line string, num int) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		logger.Debugf("#%d empty, skipping", num)
		return true
	}

	if strings.HasPrefix(trimmed, "//") {
		logger.Debugf("#%d comment, skipping", num)
		return true
	}

	return false
}
