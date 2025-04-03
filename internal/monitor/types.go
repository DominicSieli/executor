package monitor

import "time"
import "github.com/fsnotify/fsnotify"

type Config struct {
	Directory	   string
	Command		   []string
	DebouncePeriod time.Duration
}

type FileMonitor struct {
	config	*Config
	watcher *fsnotify.Watcher
}