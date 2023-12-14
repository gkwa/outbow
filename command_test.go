package outbow

import (
	"testing"
)

func TestCommandResult_Run(t *testing.T) {
	// Example usage
	command := "echo"
	args := []string{"Hello, World!"}

	result := &CommandResult{
		Command: command,
		Args:    args,
	}

	err := result.Run()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Check the result
	expectedCommand := "echo Hello, World!"
	if result.CommandString() != expectedCommand {
		t.Errorf("Expected command: %s, got: %s", expectedCommand, result.CommandString())
	}

	expectedExitCode := 0
	if result.ExitCode != expectedExitCode {
		t.Errorf("Expected exit code: %d, got: %d", expectedExitCode, result.ExitCode)
	}

	expectedStdout := "Hello, World!\n"
	if result.Stdout != expectedStdout {
		t.Errorf("Expected stdout: %s, got: %s", expectedStdout, result.Stdout)
	}

	expectedStderr := ""
	if result.Stderr != expectedStderr {
		t.Errorf("Expected stderr: %s, got: %s", expectedStderr, result.Stderr)
	}
}
