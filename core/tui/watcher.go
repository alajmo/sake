package tui

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func WatchFiles(app *App, paths ...string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					app.Reload()
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
