package hundfile

import (
	"fmt"
	"hund/cli"
	"hund/util"
)

type Target struct {
	Name   string
	Parser *cli.CliParser
	Script string
}

func NewTarget() Target {
	parser := cli.NewCliParser()
	target := Target{}
	target.Parser = parser
	return target
}

func (self Target) String() string {
	script := util.EscapeNL(self.Script)
	return fmt.Sprintf("Name: \"%s\"\nScript: \"%s\"\nParser: %s", self.Name, script, self.Parser)
}
func (self Target) Matches(name string) bool {
	return self.Name == name
}
