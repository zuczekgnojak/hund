package hundfile

import (
	"fmt"
)

type Options struct {
	ProgramName      string
	ScriptsDirectory string
	HundfileName     string
	VerboseMode      bool
	DryRun           bool
	ShowHelp         bool
}

func (self Options) String() string {
	return fmt.Sprintf(
		"ProgramName: \"%s\"\nScriptsDirectory: \"%s\"\nHundfileName: \"%s\"\nVerboseMode: %v\nDryRun: %v\nShowHelp: %v",
		self.ProgramName, self.ScriptsDirectory, self.HundfileName, self.VerboseMode, self.DryRun, self.ShowHelp,
	)
}

func NewOptions() Options {
	opt := Options{
		ProgramName:      "hund",
		ScriptsDirectory: "/tmp",
		HundfileName:     "Hundfile",
		VerboseMode:      false,
		DryRun:           false,
		ShowHelp:         false,
	}
	return opt
}

func (self Options) GetHelp() string {
	result := fmt.Sprintf("%s [options] target-name [target-options] target-args\n", self.ProgramName)
	result += "\n"
	result += "options:\n"
	result += "--filename, -f value\tpath to a Hundfile\n"
	result += "--temp-dir, -t value\tpath to a directory storing rendered script before execution\n"
	result += "--verbose, -v\t\tshow verbose information about program execution\n"
	result += "--dry-run, -d\t\trender and print script, don't run it\n"
	result += "--help, -h\t\tshow this help and exit\n"
	return result
}
