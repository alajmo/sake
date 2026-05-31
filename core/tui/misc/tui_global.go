package misc

import (
	"github.com/alajmo/sake/core/dao"
	"github.com/rivo/tview"
)

var Config *dao.Config

var App *tview.Application
var Pages *tview.Pages
var MainPage *tview.Pages
var PreviousPane tview.Primitive

// Nav buttons
var ServerBtn *tview.Button
var TaskBtn *tview.Button
var RunBtn *tview.Button
var ExecBtn *tview.Button
var HelpBtn *tview.Button

// Last focus per page
var ServersLastFocus *tview.Primitive
var TasksLastFocus *tview.Primitive
var RunLastFocus *tview.Primitive
var ExecLastFocus *tview.Primitive

// Misc
var HelpModal *tview.Modal
var Search *tview.InputField
