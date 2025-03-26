package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for monitoring mode
	fmt.Println("Choose monitoring mode:")
	fmt.Println("1. Monitor a specific file")
	fmt.Println("2. Monitor files by extension")
	fmt.Print("Enter choice (1 or 2): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var watcher *fsnotify.Watcher
	var err error
	var filesToWatch []string

	if choice == "1" {
		// Monitor single file
		fmt.Print("Enter the file path to monitor: ")
		filePath, _ := reader.ReadString('\n')
		filePath = strings.TrimSpace(filePath)

		// Verify file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Fatalf("File does not exist: %s", filePath)
		}

		filesToWatch = []string{filePath}
	} else if choice == "2" {
		// Monitor by file extension
		fmt.Print("Enter file extensions to monitor (comma separated, e.g., go,txt,json): ")
		extensionsInput, _ := reader.ReadString('\n')
		extensionsInput = strings.TrimSpace(extensionsInput)
		extensions := strings.Split(extensionsInput, ",")

		// Clean up extensions (remove dots and trim spaces)
		for i, ext := range extensions {
			ext = strings.TrimPrefix(ext, ".")
			ext = strings.TrimSpace(ext)
			extensions[i] = ext
		}

		// Find all files with matching extensions in current dir and subdirs
		err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ext := strings.TrimPrefix(filepath.Ext(path), ".")
				for _, targetExt := range extensions {
					if ext == targetExt {
						filesToWatch = append(filesToWatch, path)
						break
					}
				}
			}
			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		if len(filesToWatch) == 0 {
			log.Fatal("No files found with the specified extensions")
		}

		fmt.Printf("Monitoring %d files with extensions: %s\n", len(filesToWatch), extensionsInput)
	} else {
		log.Fatal("Invalid choice. Please enter 1 or 2.")
	}

	// Get command to execute
	fmt.Print("Enter command to execute when files change: ")
	commandInput, _ := reader.ReadString('\n')
	commandInput = strings.TrimSpace(commandInput)
	commandParts := strings.Fields(commandInput)

	// Create watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add directories containing the files to watch
	dirsAdded := make(map[string]bool)
	for _, file := range filesToWatch {
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
				for _, watchedFile := range filesToWatch {
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