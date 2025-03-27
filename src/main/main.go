package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"bufio"
	"os/exec"
	"strings"
	"path/filepath"
	"executor/src/utilities"
	"github.com/fsnotify/fsnotify"
)

func main() {
	var err error
	var watcher *fsnotify.Watcher
	defer watcher.Close()
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Directory: ")
	directoryInput, _ := reader.ReadString('\n')
	directory := strings.TrimSpace(directoryInput)

	files := utilities.FindFiles(directory)

	if len(files) == 0 {
		fmt.Println("No files found")
		os.Exit(0)
	}

	fmt.Print("Commands: ")
	commandInput, _ := reader.ReadString('\n')
	commandInput = strings.TrimSpace(commandInput)
	commandParts := strings.Fields(commandInput)

	if len(commandParts) == 0 {
		fmt.Println("No commands specified")
		os.Exit(0)
	}

	// Create watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	// Add directories containing the files to watch
	dirsAdded := make(map[string]bool)
	for _, file := range files {
		dir := filepath.Dir(file)
		if !dirsAdded[dir] {
			err = watcher.Add(dir)
			if err != nil {
				log.Printf("Error watching directory %s: %v\n", dir, err)
			} else {
				dirsAdded[dir] = true
			}
		}
	}

	// Channel to debounce events
	eventCh := make(chan struct{}, 1)

	// Start goroutine to handle events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Check if the event is for one of our watched files
				for _, watchedFile := range files {
					if event.Name == watchedFile && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
						// Debounce by sending to channel with non-blocking write
						select {
						case eventCh <- struct{}{}:
						default:
						}
						break
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	// Process debounced events
	fmt.Println("Monitoring files for changes... Press Ctrl+C to stop.")
	for range eventCh {
		// Wait a short time to catch multiple rapid changes
		time.Sleep(200 * time.Millisecond)

		// Drain any additional events that came in during the wait
		drainEvents(eventCh)

		fmt.Println("Change detected. Executing command...")
		cmd := exec.Command(commandParts[0], commandParts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("Command execution failed: %v\n", err)
		}
	}
}

func drainEvents(ch chan struct{}) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}