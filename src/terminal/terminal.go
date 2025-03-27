package terminal

import "os"
import "os/exec"

func Clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func ExecuteCommands(commands []string) error {
	cmd := exec.Command(commands[0], commands[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
