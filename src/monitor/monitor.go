package monitor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"github.com/fsnotify/fsnotify"
)

func NewFileMonitor(config *Config, watcher *fsnotify.Watcher) *FileMonitor {
	return &FileMonitor{
		config:  config,
		watcher: watcher,
	}
}

func (fm *FileMonitor) Start() error {
	// Add all subdirectories to watcher
	if err := filepath.Walk(fm.config.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := fm.watcher.Add(path); err != nil {
				log.Printf("Error watching directory %s: %v", path, err)
			} else {
				fmt.Printf("Watching directory: %s\n", path)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Channel to debounce events
	eventCh := make(chan struct{}, 1)

	// Start goroutine to handle events
	go func() {
		for {
			select {
			case event, ok := <-fm.watcher.Events:
				if !ok {
					return
				}

				// Handle new directory creation
				if event.Op&fsnotify.Create == fsnotify.Create {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						if err := fm.watcher.Add(event.Name); err == nil {
							fmt.Printf("New directory detected and added to watch list: %s\n", event.Name)
						}
					}
				}

				// Trigger on any write or create operation
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					select {
					case eventCh <- struct{}{}:
					default:
					}
				}

			case err, ok := <-fm.watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	// Process debounced events
	fmt.Printf("Monitoring directory %s for changes... Press Ctrl+C to stop.\n", fm.config.Directory)
	for range eventCh {
		time.Sleep(fm.config.DebouncePeriod)
		drainEvents(eventCh)
		fm.executeCommand()
	}

	return nil
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

func (fm *FileMonitor) executeCommand() {
	fmt.Println("Change detected. Executing command...")
	cmd := exec.Command(fm.config.Command[0], fm.config.Command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Command execution failed: %v\n", err)
	}
}