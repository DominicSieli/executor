package main

import "log"
import "executor/src/input"
import "executor/src/events"
import "executor/src/filesystem"
import "github.com/fsnotify/fsnotify"

func main() {
	watcher, err := fsnotify.NewWatcher()
	defer watcher.Close()

	if err != nil {
		log.Fatal(err)
	}

	directory := input.GetDirectory()
	commands := input.GetCommands()

	files := filesystem.FindFiles(directory)
	filesystem.AddDirectories(files, watcher)

	eventChannel := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				for _, watchedFile := range files {
					if event.Name == watchedFile && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
						// Debounce by sending to channel with non-blocking write
						select {
						case eventChannel <- struct{}{}:
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

	events.ProcessDebouncedEvents(eventChannel, commands)
}