package main

import (
	"fmt"
	"hund/hundfile"
	"hund/logger"
	"hund/parser"
	"hund/run"
	"os"
)

func main() {
	args := os.Args
	options := hundfile.NewOptions()
	args, err := parser.ParseOptions(args, &options)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.SetVerbose(options.VerboseMode)
	logger.Debugf("Options\n%s\n", options)

	if options.ShowHelp {
		fmt.Println(options.GetHelp())
		return
	}

	hundfileData, err := parser.ReadFile(options.HundfileName)
	if err != nil {
		logger.Error(err)
		return
	}

	hundfileParser := parser.NewHundfileParser()
	hundfile, err := hundfileParser.Parse(hundfileData)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Debugf("Hundfile\n%s\n", hundfile)

	renderer := run.NewRenderer(hundfile)
	script, err := renderer.Render(args)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Debugf("Script\n%s\n", script)

	if options.DryRun {
		fmt.Println(script)
		return
	}

	executor := run.NewExecutor(options, hundfile)
	logger.Debugf("Executor\n%s\n", executor)

	statusCode, err := executor.Exec(script)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Debugf("exit status %d", statusCode)
	os.Exit(statusCode)
}
