package input

import "os"
import "fmt"
import "bufio"
import "strings"
import "executor/src/terminal"

func GetDirectory() string {
	terminal.Clear()
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Directory: ")
	input, _ := reader.ReadString('\n')
	directory := strings.TrimSpace(input)

	if len(directory) == 0 {
		fmt.Println("No directory specified")
		os.Exit(0)
	}

	return directory
}

func GetCommands() []string {
	terminal.Clear()
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Commands: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	commands := strings.Fields(input)

	if len(commands) == 0 {
		fmt.Println("No commands specified")
		os.Exit(0)
	}

	return commands
}
