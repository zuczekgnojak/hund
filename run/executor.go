package run

import (
	"fmt"
	"hund/hundfile"
	"hund/logger"
	"hund/util"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

type Executor struct {
	shell     string
	shellArgs []string
	tempDir   string
}

func NewExecutor(options hundfile.Options, hundfile hundfile.Hundfile) Executor {
	return Executor{
		shell:     hundfile.Shell,
		shellArgs: hundfile.ShellArgs,
		tempDir:   options.ScriptsDirectory,
	}
}

func (e Executor) Exec(script string) (int, error) {
	f, err := os.CreateTemp(e.tempDir, "hund-run")
	if err != nil {
		return 0, err
	}
	defer f.Close()

	logger.Debugf("created script file %s\n", f.Name())

	_, err = f.WriteString(script)
	if err != nil {
		return 0, err
	}

	defer os.Remove(f.Name())

	args := append(e.shellArgs, f.Name())
	cmd := exec.Command(e.shell, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debugln("starting target")
	err = cmd.Start()
	if err != nil {
		return 0, err
	}
	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	logger.Debugln("waiting for target to finish")
	err = cmd.Wait()
	if err == nil {
		return 0, nil
	}

	exitError, ok := err.(*exec.ExitError)
	if !ok {
		return 0, err
	}

	return exitError.ExitCode(), nil
}

func (self Executor) String() string {
	args := []string{}
	for _, arg := range self.shellArgs {
		args = append(args, util.Quote(arg))
	}
	shellArgs := strings.Join(args, ", ")

	return fmt.Sprintf(
		"Shell: \"%s\"\nShellArgs: [%s]\nTempDir: \"%s\"",
		self.shell, shellArgs, self.tempDir,
	)
}
