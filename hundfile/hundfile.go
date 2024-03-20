package hundfile

import (
	"fmt"
	"hund/util"
	"strings"
)

type Hundfile struct {
	Globals   []string
	Targets   []Target
	Shell     string
	ShellArgs []string
	EmbedSep  string
	FlagValue string
}

func NewHundfile() Hundfile {
	return Hundfile{
		Shell:     "/bin/sh",
		EmbedSep:  ";",
		FlagValue: "x",
	}
}

func (self *Hundfile) AddTarget(target Target) error {
	for _, t := range self.Targets {
		if t.Matches(target.Name) {
			return util.NewError("target with \"%s\" already exists", target.Name)
		}
	}
	self.Targets = append(self.Targets, target)
	return nil
}

func (self Hundfile) GetTarget(targetName string) (Target, error) {
	for _, target := range self.Targets {
		if target.Matches(targetName) {
			return target, nil
		}
	}
	return Target{}, util.NewError("could not find target \"%s\"", targetName)
}

func (self *Hundfile) ApplyGlobal(name string, args string) error {
	switch name {
	case "shell":
		splitedArgs := util.StringToArgs(args)
		if len(splitedArgs) < 1 {
			return util.NewError("too few arguments to directive @shell")
		}
		self.Shell = splitedArgs[0]
		self.ShellArgs = splitedArgs[1:]
	case "embedSep":
		self.EmbedSep = args
	case "flagValue":
		self.FlagValue = args

	default:
		return util.NewError("invalid global directive %s", name)
	}
	return nil
}

func (self Hundfile) String() string {
	globals := strings.Join(self.Globals, ", ")

	targetNames := []string{}
	for _, target := range self.Targets {
		targetNames = append(targetNames, util.Quote(target.Name))
	}
	targets := strings.Join(targetNames, ", ")

	shellArgs := []string{}
	for _, arg := range self.ShellArgs {
		shellArgs = append(shellArgs, util.Quote(arg))
	}
	shellArgsStr := strings.Join(shellArgs, ", ")

	return fmt.Sprintf(
		"Shell: \"%s\"\nShellArgs: [%s]\nEmbedSep: \"%s\"\nFlagValue: \"%s\"\nGlobals: [%s]\nTargets: [%s]",
		self.Shell, shellArgsStr, self.EmbedSep, self.FlagValue, globals, targets,
	)
}
