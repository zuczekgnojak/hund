package run

import (
	"fmt"
	"hund/cli"
	"hund/hundfile"
	"hund/logger"
	"hund/parser"
	"hund/util"
	"strings"
)

type Renderer struct {
	hundfile       hundfile.Hundfile
	visitedTargets []string
}

func NewRenderer(hundfile hundfile.Hundfile) Renderer {
	return Renderer{
		hundfile:       hundfile,
		visitedTargets: []string{},
	}
}

func (self *Renderer) Render(args []string) (string, error) {
	if len(args) == 0 {
		return "", util.NewError("missing target name")
	}

	targetName := args[0]
	args = args[1:]

	logger.Debugf("rendering \"%s\"", targetName)

	script, err := self.innerRender(targetName, args)
	if err != nil {
		return "", err
	}

	return script, nil
}

func (self *Renderer) innerRender(targetName string, args []string) (string, error) {
	logger.Debugf("rendering script \"%s\"", targetName)
	script := ""
	if self.visited(targetName) {
		circle := strings.Join(self.visitedTargets, " -> ")
		circle += fmt.Sprintf(" -> %s", targetName)
		return script, util.NewError("detected circular dependency %s", circle)
	}
	self.visitedTargets = append(self.visitedTargets, targetName)

	target, err := self.hundfile.GetTarget(targetName)
	if err != nil {
		return script, err
	}

	logger.Debugf("parsing %d arguments", len(args))
	variables := make(map[string]string)
	writer := cli.NewMapWriter(variables, self.hundfile.FlagValue)
	leftoverArgs, err := target.Parser.Parse(args, writer)
	if err != nil {
		return script, err
	}
	if len(leftoverArgs) > 0 {
		return script, util.NewError("invalid arguments %v", leftoverArgs)
	}
	script = target.Script

	logger.Debugf("rendering variables")
	variableUses := parser.GetVariables(script)
	for _, variableUse := range variableUses {
		name := variableUse.Name
		value := variables[name]
		script = strings.ReplaceAll(script, variableUse.InScript, value)
	}

	logger.Debugf("rendering calls")
	callUses := parser.GetCalls(script)
	for _, callUse := range callUses {
		args := util.StringToArgs(callUse.Name)
		if len(args) < 1 {
			return script, util.NewError("invalid call %s", callUse.InScript)
		}
		callName := args[0]
		callArgs := args[1:]
		callResult, err := self.innerRender(callName, callArgs)
		if err != nil {
			return script, err
		}
		script = strings.ReplaceAll(script, callUse.InScript, callResult)
	}

	logger.Debugf("rendering embeds")
	embedUses := parser.GetEmbeds(script)
	for _, embedUse := range embedUses {
		args := util.StringToArgs(embedUse.Name)
		if len(args) < 1 {
			return script, util.NewError("invalid embed %s", embedUse.InScript)
		}
		embedName := args[0]
		embedArgs := args[1:]
		embedResult, err := self.innerRender(embedName, embedArgs)
		if err != nil {
			return script, err
		}
		embedSep := self.hundfile.EmbedSep
		embedResult = strings.ReplaceAll(embedResult, "\n", embedSep)
		script = strings.ReplaceAll(script, embedUse.InScript, embedResult)
	}

	self.visitedTargets = self.visitedTargets[:len(self.visitedTargets)-1]
	return script, nil
}

func (self *Renderer) visited(targetName string) bool {
	for _, name := range self.visitedTargets {
		if name == targetName {
			return true
		}
	}
	return false
}
