package cli

import (
	"fmt"
	"hund/logger"
	"hund/util"
	"regexp"
	"strings"
)

type ArgumentKind int

const (
	SingleArg     ArgumentKind = iota
	OptionalArg   ArgumentKind = iota
	AtLeastOneArg ArgumentKind = iota
	AnyArg        ArgumentKind = iota
)

type Argument struct {
	name string
	kind ArgumentKind
}

func (self Argument) isTerminating() bool {
	return self.kind != SingleArg
}

type OptionKind int

const (
	FlagOpt  OptionKind = iota
	ValueOpt OptionKind = iota
)

type Option struct {
	name      string
	shortname string
	kind      OptionKind
}

type CliParser struct {
	options        []Option
	arguments      []Argument
	terminatingArg bool
}

func (self *CliParser) String() string {
	options := []string{}
	for _, opt := range self.options {
		options = append(options, opt.name)
	}

	arguments := []string{}
	for _, arg := range self.arguments {
		arguments = append(arguments, arg.name)
	}

	opts := strings.Join(options, ", ")
	args := strings.Join(arguments, ", ")
	return fmt.Sprintf("{options: [%s], arguments: [%s]}", opts, args)
}

func NewCliParser() *CliParser {
	return &CliParser{}
}

func (self *CliParser) AddArgument(kind ArgumentKind, name string) error {
	logger.Debugf("adding argument \"%s\" to parser", name)
	if self.Contains(name) {
		return util.NewError("can't add argument, parser alread has \"%s\"", name)
	}

	if self.terminatingArg {
		lastArg := self.arguments[len(self.arguments)-1]
		return util.NewError("can't add more arguments, arg \"%s\" is terminating", lastArg.name)
	}

	argument := Argument{name: name, kind: kind}

	if argument.isTerminating() {
		self.terminatingArg = true
	}

	self.arguments = append(self.arguments, argument)
	return nil
}

func (self *CliParser) AddOption(kind OptionKind, name string, shortname ...string) error {
	logger.Debugf("adding option \"%s\" to parser", name)
	if len(shortname) > 1 {
		return util.NewError("too many values passed as shortname")
	}

	if self.Contains(name) {
		return util.NewError("can't add option, parser alread has \"%s\"", name)
	}

	short := ""
	if len(shortname) == 1 {
		short = shortname[0]
	}

	option := Option{name: name, shortname: short, kind: kind}

	for _, opt := range self.options {
		if opt.shortname != "" && opt.shortname == option.shortname {
			return util.NewError("option %s already defined", option.shortname)
		}
	}

	self.options = append(self.options, option)
	return nil
}

func (self *CliParser) Add(spec string) error {
	logger.Debugf("adding spec \"%s\"", spec)
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return util.NewError("empty spec provided")
	}

	argumentExp := regexp.MustCompile(`^(?P<name>[a-zA-Z][a-zA-Z0-9-]*)(?P<type>[\?\*\+])?$`)
	matches := argumentExp.FindStringSubmatch(spec)
	if matches != nil {
		logger.Debugf("detected argument")
		specName := matches[argumentExp.SubexpIndex("name")]
		specType := matches[argumentExp.SubexpIndex("type")]
		var kind ArgumentKind
		switch specType {
		case "?":
			kind = OptionalArg
		case "+":
			kind = AtLeastOneArg
		case "*":
			kind = AnyArg
		default:
			kind = SingleArg
		}
		return self.AddArgument(kind, specName)
	}
	logger.Debugf("spec does not represent argument")

	optionExp := regexp.MustCompile(`^(?P<name>[a-zA-Z][a-zA-Z-]*)(\|(?P<short>[a-zA-Z]))?=(?P<type>value|flag)`)
	matches = optionExp.FindStringSubmatch(spec)
	if matches == nil {
		return util.NewError("failed to parse spec \"%s\"", spec)
	}

	optName := matches[optionExp.SubexpIndex("name")]
	optShort := matches[optionExp.SubexpIndex("short")]
	optType := matches[optionExp.SubexpIndex("type")]

	kind := FlagOpt
	if optType == "value" {
		kind = ValueOpt
	}
	return self.AddOption(kind, optName, optShort)
}

func (self *CliParser) Contains(name string) bool {
	for _, argument := range self.arguments {
		if argument.name == name {
			return true
		}
	}

	for _, option := range self.options {
		if option.name == name {
			return true
		}
	}
	return false
}

func (self *CliParser) Names() []string {
	result := []string{}

	for _, option := range self.options {
		result = append(result, option.name)
	}

	for _, argument := range self.arguments {
		result = append(result, argument.name)
	}

	return result
}

func (self *CliParser) Parse(args []string, writer CliWriter) ([]string, error) {
	logger.Debugf("parsing %v", args)
	args, err := self.parseOptions(args, writer)
	if err != nil {
		return args, err
	}

	if len(self.arguments) == 0 {
		return args, nil
	}

	args, err = self.parseArguments(args, writer)
	return args, err
}

func (self *CliParser) parseOptions(args []string, writer CliWriter) ([]string, error) {
	for len(args) > 0 {
		arg := args[0]
		token, err := parseArgToken(arg)
		if err != nil {
			return args, err
		}

		if token.kind == ValueArg {
			break
		}

		option, ok := self.findOption(token)
		if !ok {
			return args, util.NewError("invalid option \"%s\"", token.value)
		}

		if option.kind == FlagOpt {
			args = args[1:]
			err = writer.Write(option.name)
			if err != nil {
				return args, err
			}

			continue
		}

		if len(args) < 2 {
			return args, util.NewError("missing value for option \"%s\"", option.name)
		}
		valToken, err := parseArgToken(args[1])
		if err != nil {
			return args, err
		}

		if valToken.kind != ValueArg {
			return args, util.NewError("invalid value \"%s\" for option \"%s\"", args[1], option.name)
		}

		err = writer.Write(option.name, valToken.value)
		if err != nil {
			return args, err
		}
		args = args[2:]
	}
	return args, nil
}

func (self *CliParser) findOption(token ArgToken) (Option, bool) {
	for _, option := range self.options {
		if token.kind == ShortOpt && option.shortname == token.value {
			return option, true
		}
		if token.kind == LongOpt && option.name == token.value {
			return option, true
		}
	}
	return Option{}, false
}

func (self *CliParser) parseArguments(args []string, writer CliWriter) ([]string, error) {
	valueTokens, err := getValueTokens(args)
	if err != nil {
		return args, err
	}

	argsConsumed := 0
	for _, argument := range self.arguments {
		if !argument.isTerminating() {
			if len(valueTokens) == 0 {
				return args, util.NewError("missing argment \"%s\"", argument.name)
			}

			writer.Write(argument.name, valueTokens[0].value)
			argsConsumed += 1
			valueTokens = valueTokens[1:]
			continue
		}

		tokensLeft := len(valueTokens)
		if argument.kind == OptionalArg && tokensLeft > 1 {
			return args, util.NewError("too many arguments, \"%s\" accepts at most 1 value", argument.name)
		}

		if argument.kind == AtLeastOneArg && tokensLeft == 0 {
			return args, util.NewError("missing argument \"%s\"", argument.name)
		}

		values := []string{}

		for _, token := range valueTokens {
			values = append(values, token.value)
		}
		argsConsumed += len(valueTokens)
		writer.Write(argument.name, strings.Join(values, " "))
	}
	return args[argsConsumed:], nil
}

func getValueTokens(args []string) ([]ArgToken, error) {
	tokens := []ArgToken{}

	for _, arg := range args {
		token, err := parseArgToken(arg)
		if err != nil {
			return tokens, err
		}
		if token.kind != ValueArg {
			break
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}
