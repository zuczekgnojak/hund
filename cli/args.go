package cli

import (
	"hund/util"
	"strings"
)

type ArgTokenKind int

const (
	ShortOpt ArgTokenKind = iota
	LongOpt  ArgTokenKind = iota
	ValueArg ArgTokenKind = iota
)

type ArgToken struct {
	value string
	kind  ArgTokenKind
}

func parseArgToken(arg string) (ArgToken, error) {
	token := ArgToken{}

	if arg == "" {
		return token, util.NewError("empty string")
	}

	token.value = strings.TrimLeft(arg, "-")
	diff := len(arg) - len(token.value)

	if diff > 2 || token.value == "" {
		return token, util.NewError("invalid argument \"%s\"", arg)
	}

	switch diff {
	case 0:
		token.kind = ValueArg
	case 1:
		token.kind = ShortOpt
	case 2:
		token.kind = LongOpt
	}

	if token.kind == ShortOpt && len(token.value) > 1 {
		return token, util.NewError("invalid short option %s", arg)
	}

	return token, nil
}
