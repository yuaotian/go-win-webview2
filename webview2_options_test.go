//go:build windows
// +build windows

package webview2

import "testing"

func TestNormalizeWindowOptionsUsesDefaultsForZeroValue(t *testing.T) {
	defaults := DefaultWindowOptions()
	got := normalizeWindowOptions(WindowOptions{})

	if got.Title != defaults.Title {
		t.Fatalf("expected default title %q, got %q", defaults.Title, got.Title)
	}
	if got.Width != defaults.Width {
		t.Fatalf("expected default width %d, got %d", defaults.Width, got.Width)
	}
	if got.Height != defaults.Height {
		t.Fatalf("expected default height %d, got %d", defaults.Height, got.Height)
	}
	if got.Resizable != defaults.Resizable {
		t.Fatalf("expected default resizable %v, got %v", defaults.Resizable, got.Resizable)
	}
}

func TestNormalizeWindowOptionsKeepsExplicitFalse(t *testing.T) {
	got := normalizeWindowOptions(WindowOptions{
		Title:     "custom",
		Width:     1024,
		Height:    768,
		Resizable: false,
	})

	if got.Resizable {
		t.Fatalf("expected explicit false to be preserved")
	}
	if got.Title != "custom" {
		t.Fatalf("expected custom title to be preserved")
	}
	if got.Width != 1024 || got.Height != 768 {
		t.Fatalf("expected explicit size to be preserved, got %d x %d", got.Width, got.Height)
	}
}
