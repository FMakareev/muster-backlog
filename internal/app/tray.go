package app

import (
	"runtime"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/FMakareev/muster-backlog/internal/settings"
)

// MainWindowName is the window the tray shows and hides.
const MainWindowName = "muster.main"

var (
	trayMu sync.Mutex
	// tray is the system tray icon while tray behaviour is on, nil otherwise.
	tray *application.SystemTray
	// trayMode is read by the window's closing handler, which is installed
	// once and then follows the preference rather than being rewired.
	trayMode bool
	// trayUnavailable records that the desktop has nowhere to put an icon.
	trayUnavailable bool
	// markPNG is the application's own icon, handed in from main so the
	// picture lives with the build assets rather than being duplicated here.
	// Empty in tests and in server mode, where nothing draws it.
	markPNG []byte
)

// SetMark gives the application its icon.
//
// A tray item with no icon is a blank space with a menu behind it, which is
// exactly what this was until the mark existed. Called once from main, before
// anything can create a tray.
func SetMark(png []byte) {
	trayMu.Lock()
	markPNG = png
	trayMu.Unlock()
}

// applyWindowBehaviour turns the tray on or off to match the preference.
//
// It is safe to call before the application exists - during tests, or in
// server mode - and simply does nothing then.
func applyWindowBehaviour(prefs settings.Settings) {
	app := application.Get()
	if app == nil {
		return
	}

	want := prefs.OnWindowClose == settings.BehaviourTray

	trayMu.Lock()
	defer trayMu.Unlock()

	if want && !TrayAvailable() {
		// A desktop with no status-notifier host has nowhere to put an icon.
		// Falling back to ordinary window behaviour is the only option that
		// does not make the application vanish with no way to get it back.
		trayUnavailable = true
		trayMode = false
		return
	}
	trayUnavailable = false

	switch {
	case want && tray == nil:
		tray = newTray(app)
		trayMode = true
	case !want && tray != nil:
		tray.Destroy()
		tray = nil
		trayMode = false
	default:
		trayMode = want
	}
}

// TrayUnavailable reports that tray behaviour was asked for and could not be
// provided, so the interface can say so rather than leave the setting looking
// as though it worked.
func TrayUnavailable() bool {
	trayMu.Lock()
	defer trayMu.Unlock()
	return trayUnavailable
}

// newTray builds the icon and its menu.
func newTray(app *application.App) *application.SystemTray {
	t := app.SystemTray.New()
	t.SetLabel("Muster")

	// The size a tray actually draws is the reason the mark is a letter built
	// from blocks rather than an arrangement of small parts: it is about
	// twenty-two pixels here.
	trayMu.Lock()
	mark := markPNG
	trayMu.Unlock()
	if len(mark) > 0 {
		t.SetIcon(mark)
	}

	menu := application.NewMenu()
	menu.Add("Show Muster").OnClick(func(*application.Context) {
		showMainWindow()
	})
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(*application.Context) {
		trayMu.Lock()
		trayMode = false
		trayMu.Unlock()
		app.Quit()
	})
	t.SetMenu(menu)

	t.OnClick(func() { showMainWindow() })
	return t
}

func showMainWindow() {
	app := application.Get()
	if app == nil {
		return
	}
	if window, ok := app.Window.GetByName(MainWindowName); ok {
		window.Show()
		window.Focus()
	}
}

// InstallCloseBehaviour makes the window hide rather than close while tray
// behaviour is on. It is installed once at startup; the preference is read at
// the moment of closing, so changing it takes effect without rewiring.
func InstallCloseBehaviour(window *application.WebviewWindow) {
	window.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
		trayMu.Lock()
		resident := trayMode
		trayMu.Unlock()

		if resident {
			event.Cancel()
			window.Hide()
		}
	})
}

// TrayAvailable reports whether the desktop can show a tray icon.
//
// On Linux the tray is not part of the windowing system: it is a
// status-notifier host on the session bus, which many desktops do not run.
// Asking before creating the icon is the difference between a preference that
// works and an application that disappears.
func TrayAvailable() bool {
	if runtime.GOOS != "linux" {
		return true
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		return false
	}

	var names []string
	if err := conn.BusObject().Call(
		"org.freedesktop.DBus.ListNames", 0).Store(&names); err != nil {
		return false
	}
	for _, name := range names {
		if name == "org.kde.StatusNotifierWatcher" ||
			name == "org.freedesktop.StatusNotifierWatcher" {
			return true
		}
	}
	return false
}
