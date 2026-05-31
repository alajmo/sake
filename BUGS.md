# TUI bugs (`tui` branch)

Defects in the `core/tui/**` implementation and the `TextTUI` backend it calls (`core/run/text.go`), ordered by impact.

## ✅ 1. Config reload mutates the UI from non-UI goroutines - HIGH

**Files:** `core/tui/tui.go:40-50`, `core/tui/tui_input.go:37-39`, `core/tui/watcher.go:18-25`

`App.Reload()` rebuilds `misc.Pages` and calls `app.App.SetRoot(...)` + `app.App.Draw()`. It is invoked off the event loop from two places: F5 runs `go app.Reload()`, and the fsnotify watcher calls `app.Reload()` directly from its goroutine. tview is single-threaded — widget mutation must go through `App.QueueUpdate`/`QueueUpdateDraw`. As written, a reload can run concurrently with the event loop and with a second reload, racing on `misc.Pages`, the input handler, and the screen (data race, possible panic / corruption).

**Fix:** Run the reload body inside `app.App.QueueUpdateDraw(...)` and serialize reloads (debounce the watcher).

## ✅ 2. `EventEmitter` runs listeners in goroutines that mutate widgets - HIGH

**Files:** `core/tui/misc/tui_event.go:31-39`, `core/tui/pages/tui_run.go:135`, `core/tui/pages/tui_exec.go:182`, `core/tui/pages/tui_server.go:82`

`Publish` dispatches each listener with `go listener(event)`. The "clear all" (`C`) handlers call `Emitter.Publish(filter_servers)`, so `filterServers()` runs on a detached goroutine and calls `Table.Clear()`/`SetCell`/tree rebuild concurrently with drawing → data race. (`PublishAndWait` blocks the caller so it is borderline-safe, but still executes listeners off the event loop.)

**Fix:** Make the emitter synchronous, or marshal UI mutations through `QueueUpdateDraw`.

## ✅ 3. `FocusNext`/`FocusPrevious` nil-pointer panic and wrong target - HIGH

**File:** `core/tui/misc/tui_focus.go:13-67`

If neither the current focus nor `PreviousPane` is found in `elements`, `nextFocusItem` stays a zero `TItem{}` and `nextFocusItem.Box.SetBorderColor(...)` dereferences a nil `*Box` → panic (reachable via Tab/Shift-Tab right after a page switch). Separately, `FocusPrevious`'s second loop runs unconditionally (it lacks the `if prevIndex < 0` guard that `FocusNext` has), so when both the current focus and `PreviousPane` are in the list it overrides the correct result → Shift-Tab focuses the wrong pane.

**Fix:** Guard the `PreviousPane` fallback behind "not found", and bail out when no match exists.

## ✅ 4. Output never streams; it appears only after the whole run finishes - MEDIUM

**Files:** `core/tui/components/tui_output.go`, `core/tui/pages/tui_run.go:340-345`, `core/tui/pages/tui_exec.go:252-255`

The Output `TextView` has no `SetChangedFunc`, and tview's `TextView.write()` only repaints via that callback. The single redraw is `misc.App.QueueUpdateDraw(func(){})` emitted *after* the run loop completes. For any long-running task the Output pane stays blank/frozen until the run is done.

**Fix:** Set a (throttled) changed func that calls `App.Draw()`/`QueueUpdateDraw` so output renders as it arrives.

## ✅ 5. Tag filter produces duplicate servers - MEDIUM

**File:** `core/tui/views/tui_server_view.go:469-484`

With ≥2 tags selected, the inner `break` only exits the `serverTag` loop, so a server carrying multiple selected tags is appended once per matching tag → duplicate rows in the table and tree, and duplicated execution targets.

**Fix:** Continue to the next server after the first tag match (label the `tag` loop, or dedupe with a `seen` set).

## ✅ 6. Server-tree group order is nondeterministic - MEDIUM

**File:** `core/tui/views/tui_server_view.go:265-323`

`getServerTreeHierarchy` ranges over the `groups` map, so group ordering reshuffles randomly on every filter/redraw.

**Fix:** Collect the group keys, sort them, then build nodes in sorted order.

## ✅ 7. Re-entrant runs are not guarded - MEDIUM

**Files:** `core/tui/pages/tui_run.go:316-346`, `core/tui/pages/tui_exec.go:228-256`

Pressing Ctrl-r again mid-run spawns a second execution goroutine sharing the same Output writer and opening a second set of SSH clients; output interleaves and clients can collide. The spec fields (`spec.IgnoreErrors`, etc.) are also read from the run goroutine while the options modal can toggle them → data race.

**Fix:** Add a "running" guard that ignores Ctrl-r while a run is in flight, and snapshot the spec before launching.

## ✅ 8. `--reload` quits the app on a transient bad config - MEDIUM

**File:** `core/tui/tui.go:41-45`

On `configErr != nil`, `Reload()` calls `app.App.Stop()`. The first momentarily-invalid save while editing the watched file kills the TUI, defeating the purpose of live reload.

**Fix:** Surface the parse error (e.g., a modal/status line) and keep the previous config loaded.

## ✅ 10. Table rendering measures bytes but pads/truncates as runes - LOW

**Files:** `core/tui/pages/tui_run.go:466-535`, `core/tui/pages/tui_exec.go:457-526`

`writeTableOutput` computes widths with `len()` (bytes) but pads with `%-*s` (runes) and truncates via byte slicing `cellContent[:colWidths[i]-3]`, so non-ASCII output misaligns and can be cut mid-rune. Same byte-vs-rune issue in `getTextModalSize` (`tui_modal.go:171`) and `GetModalSize` (`misc/tui_utils.go:31`).

**Fix:** Use `utf8.RuneCountInString` / `[]rune` slicing (or a width-aware helper) for measurement and truncation.

## ⬜ 11. Editing via `o` does not refresh the in-memory config - LOW

**Files:** `core/tui/views/tui_server_view.go:587-594`, `core/tui/views/tui_task_view.go:329-336`

`editServer`/`editTask` only `Suspend` + edit the file. Without `--reload` the change is invisible, and the in-memory config used for subsequent runs stays stale.

**Fix:** Re-read the config (or trigger the reload path) after the editor exits.

## ⬜ 12. `isIPAddress` accepts non-IPs - LOW

**File:** `core/tui/views/tui_server_view.go:424-435`

`fmt.Sscanf(part, "%d", …)` treats `1abc` as `1` and there is no 0–255 range check, so `1abc.2.3.4` and `999.999.999.999` are classified as IPs. Only affects the tree-grouping heuristic.

**Fix:** Use `net.ParseIP`.

## ⬜ 13. Filter case-inconsistency - LOW

**File:** `core/tui/views/tui_server_view.go:507-516`

Server/task name filters are case-insensitive, but the tag filter (`filterTags`) uses case-sensitive `strings.Contains`.

**Fix:** Lower-case both sides in `filterTags`.

## ⬜ 14. `sshServer` swallows errors and does not expand `~` - LOW

**File:** `core/tui/views/tui_server_view.go:596-651`

`cmd.Run()`'s error is ignored, and `IdentityFile` is passed unexpanded, so a `~/...` path may not resolve.

**Fix:** Surface the error and expand `~` (reuse the resolver used elsewhere in the codebase).

## ⬜ 15. Single `misc.Search` primitive shared across four pages - LOW

**Files:** `core/tui/pages.go`, all page constructors

The same `misc.Search` `InputField` is added to four different page `Flex` layouts. It works only because one page draws at a time; it is fragile tview usage (shared rect/focus).

**Fix:** Give each page its own search input, or host a single search bar above the page container.

## ⬜ 16. `wg.Add(1)` after goroutine launch - LOW

**File:** `core/run/text.go:1465`, `:1500` (TUI path); pre-existing at `:639`, `:676`

`go func(){ defer wg.Done(); … }(); wg.Add(1)` calls `Add` after starting the goroutine; `Add` must precede it, otherwise `Wait` can race to a negative-counter panic. Practically rare because the goroutine blocks on I/O first, but still incorrect.

**Fix:** Call `wg.Add(1)` before each `go`.

## Not listed

- Outside the TUI proper, the `tui` branch's import dedup in `core/dao/import_config.go:431-443` silently drops later same-named servers from imports (a behavior change) and is O(n²).
