package main

import "os"
import "fmt"
import "log"
import "time"
import "bufio"
import "strings"
import "path/filepath"
import "executor/src/monitor"
import "executor/src/terminal"
import "github.com/fsnotify/fsnotify"

func main() {
	reader := bufio.NewReader(os.Stdin)

	terminal.Clear()
	fmt.Print("Directory: ")
	directory, _ := reader.ReadString('\n')
	directory = strings.TrimSpace(directory)

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		terminal.Clear()
		log.Fatalf("Directory does not exist: %s", directory)
	}

	path, err := filepath.Abs(directory)

	if err != nil {
		terminal.Clear()
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	terminal.Clear()
	fmt.Print("Command: ")
	commandInput, _ := reader.ReadString('\n')
	commandInput = strings.TrimSpace(commandInput)
	command := strings.Fields(commandInput)

	config := &monitor.Config{
		Directory:		path,
		Command:		command,
		DebouncePeriod: 200 * time.Millisecond,
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		terminal.Clear()
		log.Fatal(err)
	}

	defer watcher.Close()

	monitor := monitor.NewFileMonitor(config, watcher)

	if err := monitor.Start(); err != nil {
		terminal.Clear()
		log.Fatal(err)
	}
}