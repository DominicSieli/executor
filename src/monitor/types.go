package monitor

import (
	"time"
	"github.com/fsnotify/fsnotify"
)

type Config struct {
	Directory	   string	// Directory to monitor
	Command		   []string // Command to execute
	DebouncePeriod time.Duration
}

type FileMonitor struct {
	config	*Config
	watcher *fsnotify.Watcher
}