package main

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The icon is the one place where "the tool was run" is routinely mistaken
// for "the right picture came out". Both the Windows .ico and the macOS .icns
// are containers built by a generator from our PNG, and nothing downstream
// would notice if they still held someone else's artwork - which is exactly
// the state this project shipped in until the mark existed.

// The application carries its own icon, at the size a desktop draws it.
func TestTheBinaryCarriesTheMark(t *testing.T) {
	if len(mark) == 0 {
		t.Fatal("the embedded mark is empty")
	}
	config, err := png.DecodeConfig(bytes.NewReader(mark))
	if err != nil {
		t.Fatalf("the embedded mark is not a PNG: %v", err)
	}
	// Small enough not to be resampled hard on every draw, large enough for a
	// window switcher. A tray draws this at about twenty-two pixels.
	if config.Width != 256 || config.Height != 256 {
		t.Errorf("the embedded mark is %dx%d, want 256x256", config.Width, config.Height)
	}
}

// The Windows icon has to carry the small sizes, because those are the ones a
// taskbar and a title bar draw; a container holding only 256 looks correct in
// a file listing and wrong everywhere it is used.
func TestTheWindowsIconCarriesTheSmallSizes(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("build", "windows", "icon.ico"))
	if err != nil {
		t.Fatalf("read icon.ico: %v", err)
	}
	if len(data) < 6 {
		t.Fatal("icon.ico is too short to be an icon")
	}
	reserved := binary.LittleEndian.Uint16(data[0:])
	kind := binary.LittleEndian.Uint16(data[2:])
	count := int(binary.LittleEndian.Uint16(data[4:]))
	if reserved != 0 || kind != 1 {
		t.Fatalf("icon.ico is not an .ico: reserved %d, type %d", reserved, kind)
	}

	sizes := map[int]bool{}
	for i := range count {
		at := 6 + i*16
		if at+16 > len(data) {
			t.Fatalf("entry %d runs past the end of the file", i)
		}
		width := int(data[at])
		if width == 0 {
			width = 256 // zero means 256 in the format
		}
		sizes[width] = true
	}
	for _, want := range []int{16, 32, 48, 256} {
		if !sizes[want] {
			t.Errorf("icon.ico has no %dpx member; it has %v", want, sizes)
		}
	}
}

// Both macOS containers have to hold real images rather than be empty shells.
func TestTheMacIconsCarryImages(t *testing.T) {
	for _, name := range []string{"icons.icns", "dmg-file-icon.icns"} {
		data, err := os.ReadFile(filepath.Join("build", "darwin", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !bytes.HasPrefix(data, []byte("icns")) {
			t.Errorf("%s is not an .icns", name)
			continue
		}
		members := 0
		for at := 8; at+8 <= len(data); {
			size := int(binary.BigEndian.Uint32(data[at+4:]))
			if size < 8 || at+size > len(data) {
				break
			}
			if bytes.HasPrefix(data[at+8:at+size], []byte("\x89PNG\r\n\x1a\n")) {
				members++
			}
			at += size
		}
		if members == 0 {
			t.Errorf("%s carries no PNG at all", name)
		}
	}
}

// Comments are not shipped metadata: nfpm and the desktop entry read values,
// and the warning in nfpm.yaml names the framework and its defaults on
// purpose, because naming them is the whole use of it.
func withoutComments(path string, content []byte) string {
	// build/linux/desktop is the .deb's template and has no extension, so it
	// is matched by name as well.
	switch {
	case strings.EqualFold(filepath.Base(path), "desktop"):
	case map[string]bool{".yaml": true, ".yml": true, ".desktop": true,
		".nsh": true}[strings.ToLower(filepath.Ext(path))]:
	default:
		return string(content)
	}
	var kept []string
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// Nothing of the framework's branding should be left in what ships.
//
// These two keep the name and are meant to: they are Wails' own build
// tooling, referenced by exactly these names from the Taskfile and from
// project.nsi, and neither carries a picture or a name a user ever sees.
// Renaming them breaks the Windows build and gains nothing.
var frameworkTooling = map[string]bool{
	"build/windows/nsis/wails_tools.nsh": true,
	"build/windows/wails.exe.manifest":   true,
}

func TestNoFrameworkBrandingIsShipped(t *testing.T) {
	readable := map[string]bool{
		".svg": true, ".json": true, ".desktop": true, ".txt": true,
		".xml": true, ".plist": true, ".manifest": true, ".yaml": true,
	}

	err := filepath.WalkDir("build", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		// The AppImage staging directory is build output, not a source asset.
		slashed := filepath.ToSlash(path)
		if strings.Contains(slashed, "appimage/build/") || frameworkTooling[slashed] {
			return nil
		}
		if strings.Contains(strings.ToLower(entry.Name()), "wails") {
			t.Errorf("%s is named after the framework", slashed)
			return nil
		}
		if !readable[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(strings.ToLower(withoutComments(path, content)), "wails") {
			t.Errorf("%s still names the framework", slashed)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk build: %v", err)
	}
}

// The placeholders the template ships with were committed and shipped once
// already: a .deb whose desktop entry called the product "My Product", and a
// bundle identifier of com.example.muster.
func TestNoTemplatePlaceholdersAreShipped(t *testing.T) {
	placeholders := []string{"My Product", "com.example", "My Company", "wails.io"}
	files := []string{
		"build/linux/desktop",
		"build/linux/muster.desktop",
		"build/linux/nfpm/nfpm.yaml",
		"build/darwin/Info.plist",
		"build/windows/info.json",
	}
	for _, name := range files {
		content, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		values := withoutComments(name, content)
		for _, placeholder := range placeholders {
			if strings.Contains(values, placeholder) {
				t.Errorf("%s still carries the placeholder %q", name, placeholder)
			}
		}
	}
}
