package tui

import (
	"os"

	"github.com/alajmo/sake/core"
	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/tui/components"
	"github.com/alajmo/sake/core/tui/misc"
	"github.com/rivo/tview"
)

func RunTui(config *dao.Config, reload bool) {
	app := NewApp(config)

	if reload {
		WatchFiles(app, config.Path)
	}

	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}

type App struct {
	App  *tview.Application
	Path string
}

func NewApp(config *dao.Config) *App {
	app := &App{
		App:  tview.NewApplication(),
		Path: config.Path,
	}
	app.setupApp(config)

	return app
}

func (app *App) Run() error {
	return app.App.SetRoot(misc.Pages, true).EnableMouse(true).Run()
}

func (app *App) Reload() {
	// Read + parse the config off the event loop so disk I/O does not block the UI.
	config, configErr := dao.ReadConfig(app.Path, "", "", false)
	if configErr != nil {
		// A transient bad save while editing the watched file must not kill the
		// TUI. Surface the parse error and keep the previously loaded config.
		// The error is ANSI-colorized; strip it (a tview TextView renders tview
		// tags, not ANSI) and escape stray brackets so it shows as plain text.
		errText := tview.Escape(core.Strip(configErr.Error()))
		app.App.QueueUpdateDraw(func() {
			components.OpenTextModal("reload-error-modal", errText, "Config Reload Error")
		})
		return
	}

	// tview is single-threaded: all widget mutation (rebuilding misc.Pages,
	// re-registering the input capture, SetRoot) must run on the event loop.
	// QueueUpdateDraw is safe to call from F5's goroutine and the watcher goroutine.
	app.App.QueueUpdateDraw(func() {
		app.setupApp(&config)
		app.App.SetRoot(misc.Pages, true)
	})
}

func (app *App) setupApp(config *dao.Config) {
	misc.Config = config

	// Load styles
	misc.LoadStyles()
	misc.SetupStyles()

	// Data
	servers := config.Servers
	tasks := config.Tasks
	serverTags := config.GetTags()

	// Create pages
	misc.App = app.App
	misc.Pages = createPages(servers, serverTags, tasks)

	// Global input handling
	HandleInput(app)
}
