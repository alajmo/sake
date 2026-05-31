package tui

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

// reloadDebounce collapses the burst of fsnotify Write events that a single
// file save typically emits into one reload.
const reloadDebounce = 100 * time.Millisecond

func WatchFiles(app *App, paths ...string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer watcher.Close()
		var debounce *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					// Reset the timer on every write so rapid bursts collapse
					// into a single Reload once writes settle. Reload itself
					// marshals onto the event loop via QueueUpdateDraw.
					if debounce != nil {
						debounce.Stop()
					}
					debounce = time.AfterFunc(reloadDebounce, app.Reload)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	for _, path := range paths {
		err = watcher.Add(path)
		if err != nil {
			log.Fatal(err)
		}
	}
}
