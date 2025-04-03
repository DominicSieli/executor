package monitor

import "os"
import "fmt"
import "log"
import "time"
import "os/exec"
import "path/filepath"
import "executor/src/terminal"
import "github.com/fsnotify/fsnotify"

func NewFileMonitor(config *Config, watcher *fsnotify.Watcher) *FileMonitor {
	return &FileMonitor{
		config:  config,
		watcher: watcher,
	}
}

func (fileMonitor *FileMonitor) Start() error {
	if err := filepath.Walk(fileMonitor.config.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err := fileMonitor.watcher.Add(path); err != nil {
				terminal.Clear()
				log.Printf("Error watching directory %s: %v", path, err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	eventChannel := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case event, ok := <-fileMonitor.watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						if err := fileMonitor.watcher.Add(event.Name); err == nil {
							terminal.Clear()
							fmt.Printf("New directory added: %s\n", event.Name)
						}
					}
				}

				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					select {
					case eventChannel <- struct{}{}:
					default:
					}
				}

			case err, ok := <-fileMonitor.watcher.Errors:
				if !ok {
					return
				}

				terminal.Clear()
				log.Println("Error:", err)
			}
		}
	}()

	for range eventChannel {
		time.Sleep(fileMonitor.config.DebouncePeriod)
		drainEvents(eventChannel)
		fileMonitor.executeCommand()
	}

	return nil
}

func drainEvents(channel chan struct{}) {
	for {
		select {
		case <-channel:
		default:
			return
		}
	}
}

func (fileMonitor *FileMonitor) executeCommand() {
	terminal.Clear()
	cmd := exec.Command(fileMonitor.config.Command[0], fileMonitor.config.Command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		terminal.Clear()
		log.Printf("Command execution failed: %v\n", err)
	}
}