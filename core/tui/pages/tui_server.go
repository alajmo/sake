package pages

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/sake/core/dao"
	"github.com/alajmo/sake/core/tui/misc"
	"github.com/alajmo/sake/core/tui/views"
)

type TServerPage struct {
	focusable  []*misc.TItem
	serverData *views.TServer
}

func CreateServersPage(
	servers []dao.Server,
	serverTags []string,
) *tview.Flex {
	p := &TServerPage{}

	// Data
	p.serverData = views.CreateServersData(
		servers,
		serverTags,
		[]string{"Server", "Host", "User", "Tags", "Description"},
		1,
		true,
		true,
		false,
		len(serverTags) > 0,
	)

	// Shortcut info view
	infoView := views.CreateServersInfoView()

	// Server view with table/tree toggle
	serverViewPages := p.createServerViewPages()

	// Tags panel
	serverPage := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(serverViewPages, 0, 1, true)

	if p.serverData.TagView != nil && len(serverTags) > 0 {
		serverPage.AddItem(p.serverData.TagView.Root, 25, 0, false)
	}

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(serverPage, 0, 1, true).
		AddItem(infoView, 1, 0, false).
		AddItem(misc.Search, 1, 0, false)

	// Focusable
	p.focusable = p.updateServerFocusable()
	misc.ServersLastFocus = &p.focusable[0].Primitive

	// Shortcuts
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyTab:
			nextPrimitive := misc.FocusNext(p.focusable)
			misc.ServersLastFocus = nextPrimitive
			return nil
		case tcell.KeyBacktab:
			nextPrimitive := misc.FocusPrevious(p.focusable)
			misc.ServersLastFocus = nextPrimitive
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'C': // Clear filters
				p.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_tag_filter", Data: ""})
				p.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_tag_selections", Data: ""})
				p.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_server_filter", Data: ""})
				p.serverData.Emitter.PublishAndWait(misc.Event{Name: "remove_server_selections", Data: ""})
				p.serverData.Emitter.Publish(misc.Event{Name: "filter_servers", Data: ""})
				return nil
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				misc.FocusPage(event, p.focusable)
				return nil
			}
		}
		return event
	})

	return page
}

func (p *TServerPage) createServerViewPages() *tview.Flex {
	isTable := p.serverData.ServerStyle == "server-table"

	pages := tview.NewPages().
		AddPage("server-table", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(p.serverData.ServerTableView.Root, 0, 1, true), true, isTable).
		AddPage("server-tree", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(p.serverData.ServerTreeView.Root, 0, 1, false), true, !isTable)

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true)

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlE:
			if p.serverData.ServerStyle == "server-table" {
				p.serverData.ServerStyle = "server-tree"
			} else {
				p.serverData.ServerStyle = "server-table"
			}
			pages.SwitchToPage(p.serverData.ServerStyle)
			p.focusable = p.updateServerFocusable()
			misc.App.SetFocus(p.focusable[0].Primitive)
			misc.ServersLastFocus = &p.focusable[0].Primitive
			return nil
		}
		return event
	})

	return page
}

func (p *TServerPage) updateServerFocusable() []*misc.TItem {
	var focusable []*misc.TItem

	if p.serverData.ServerStyle == "server-table" {
		focusable = append(focusable, misc.GetTUIItem(
			p.serverData.ServerTableView.Table,
			p.serverData.ServerTableView.Table.Box,
		))
	} else {
		focusable = append(focusable, misc.GetTUIItem(
			p.serverData.ServerTreeView.Tree,
			p.serverData.ServerTreeView.Tree.Box,
		))
	}

	if p.serverData.TagView != nil && len(p.serverData.ServerTags) > 0 {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				p.serverData.TagView.List,
				p.serverData.TagView.List.Box,
			),
		)
	}

	return focusable
}
