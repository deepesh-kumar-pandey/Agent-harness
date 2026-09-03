package shell

import (
	"fmt"
	"os/exec"
)

type Shell struct{}

func (s *Shell) Execute(command string, args ...string) error {

	_, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("command not found: %w", err)
	}
	cmd := exec.Command(command, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	fmt.Printf("Command executed: %s\n", output)
	return nil
}
