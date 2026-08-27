package app

import (
	"sync"
	"testing"
	"time"
)

// The mark must not be read through the tray's own lock.
//
// applyWindowBehaviour takes trayMu and, still holding it, calls newTray.
// A sync.Mutex is not reentrant, so a second Lock inside newTray hangs the
// goroutine forever with the lock held: the asset server is already up, the
// window is never created, and nothing is logged. "The server starts but no
// window opens" — and only when the tray preference is on, which is why it
// survived every test and every demo run.
//
// This exercises the same order without needing an application to exist.
func TestTheMarkIsReadableWhileTheTrayLockIsHeld(t *testing.T) {
	SetMark([]byte("a pretend icon"))

	done := make(chan []byte, 1)
	go func() {
		// Exactly what applyWindowBehaviour does before calling newTray.
		trayMu.Lock()
		defer trayMu.Unlock()
		done <- trayIcon()
	}()

	select {
	case got := <-done:
		if string(got) != "a pretend icon" {
			t.Errorf("the mark read back as %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("reading the mark deadlocked against the tray lock, which is what newTray does at startup")
	}
}

// SetMark is called from main before anything can build a tray, but a
// preference change rebuilds one on another goroutine, so the two still have
// to be safe together.
func TestTheMarkSurvivesConcurrentUse(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if i%2 == 0 {
				SetMark([]byte("icon"))
				return
			}
			trayMu.Lock()
			defer trayMu.Unlock()
			_ = trayIcon()
		}()
	}
	wg.Wait()
}
