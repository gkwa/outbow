package outbow

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type CommandResult struct {
	Command  string
	Args     []string
	Stdout   string
	Stderr   string
	ExitCode int
}

func (cr *CommandResult) CommandString() string {
	return fmt.Sprintf("%s %s", cr.Command, strings.Join(cr.Args, " "))
}

func (cr *CommandResult) Run() error {
	cmd := exec.Command(cr.Command, cr.Args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	cr.Stdout = stdout.String()
	cr.Stderr = stderr.String()

	// Check if there was an error running the command
	if err != nil {
		// Extract the exit code if available
		if exitError, ok := err.(*exec.ExitError); ok {
			cr.ExitCode = exitError.ExitCode()
		} else {
			// Set a non-zero exit code if there was a generic error
			cr.ExitCode = 1
		}
	}

	return err
}
