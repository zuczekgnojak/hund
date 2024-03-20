package parser

import (
	"errors"
	"hund/cli"
	"hund/hundfile"
)

func ParseOptions(args []string, target *hundfile.Options) ([]string, error) {
	if len(args) == 0 {
		return args, errors.New("too few arguments to parse hund options")
	}

	target.ProgramName = args[0]
	args = args[1:]

	if len(args) == 0 {
		return args, nil
	}

	cliParser := cli.NewCliParser()
	cliParser.AddOption(cli.ValueOpt, "temp-dir", "t")
	cliParser.AddOption(cli.ValueOpt, "filename", "f")
	cliParser.AddOption(cli.FlagOpt, "verbose", "v")
	cliParser.AddOption(cli.FlagOpt, "dry-run", "d")
	cliParser.AddOption(cli.FlagOpt, "help", "h")

	pointerWriter := cli.NewPointerWriter()
	pointerWriter.AddValue("temp-dir", &target.ScriptsDirectory)
	pointerWriter.AddValue("filename", &target.HundfileName)
	pointerWriter.AddFlag("verbose", &target.VerboseMode)
	pointerWriter.AddFlag("dry-run", &target.DryRun)
	pointerWriter.AddFlag("help", &target.ShowHelp)

	return cliParser.Parse(args, pointerWriter)
}
