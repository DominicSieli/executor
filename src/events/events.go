package events

import "log"
import "time"
import "executor/src/terminal"

func Drain(ch chan struct{}) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func ProcessDebouncedEvents(eventChannel chan struct{}, commands []string) {
	for range eventChannel {
		time.Sleep(200 * time.Millisecond)
		Drain(eventChannel)

		terminal.Clear()
		err := terminal.ExecuteCommands(commands)
		if err != nil {
			log.Printf("Command execution failed: %v\n", err)
		}
	}
}
