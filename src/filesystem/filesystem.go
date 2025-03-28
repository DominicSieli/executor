package filesystem

import "os"
import "fmt"
import "log"
import "path/filepath"
import "github.com/fsnotify/fsnotify"

func AddDirectories(files []string, watcher *fsnotify.Watcher) {
	directoriesAdded := make(map[string]bool)

	for _, file := range files {
		directory := filepath.Dir(file)
		if !directoriesAdded[directory] {
			err := watcher.Add(directory)
			if err != nil {
				log.Printf("Error watching directory %s: %v\n", directory, err)
			} else {
				directoriesAdded[directory] = true
			}
		}
	}
}

func FindFiles(directory string) []string {
	var files []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		fmt.Println("No files found")
		os.Exit(0)
	}

	return files
}
