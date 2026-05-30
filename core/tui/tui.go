package tui

import (
	"os"

	"github.com/alajmo/sake/core/dao"
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
	App *tview.Application
}

func NewApp(config *dao.Config) *App {
	app := &App{
		App: tview.NewApplication(),
	}
	app.setupApp(config)

	return app
}

func (app *App) Run() error {
	return app.App.SetRoot(misc.Pages, true).EnableMouse(true).Run()
}

func (app *App) Reload() {
	config, configErr := dao.ReadConfig(misc.Config.Path, "", "", false)
	if configErr != nil {
		app.App.Stop()
		return
	}

	app.setupApp(&config)
	app.App.SetRoot(misc.Pages, true)
	app.App.Draw()
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
