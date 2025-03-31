package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
	"path/filepath"
	"strings"
	"github.com/fsnotify/fsnotify"
	"executor/src/monitor"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Get directory to monitor
	fmt.Print("Enter directory to monitor: ")
	dirPath, _ := reader.ReadString('\n')
	dirPath = strings.TrimSpace(dirPath)

	// Verify directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", dirPath)
	}

	// Get absolute path
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Get command to execute
	fmt.Print("Enter command to execute when files change: ")
	commandInput, _ := reader.ReadString('\n')
	commandInput = strings.TrimSpace(commandInput)
	commandParts := strings.Fields(commandInput)

	// Create and start monitor
	config := &monitor.Config{
		Directory:		absPath,
		Command:		commandParts,
		DebouncePeriod: 200 * time.Millisecond,
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	monitor := monitor.NewFileMonitor(config, watcher)
	if err := monitor.Start(); err != nil {
		log.Fatal(err)
	}
}