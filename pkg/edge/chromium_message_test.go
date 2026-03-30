//go:build windows
// +build windows

package edge

import "testing"

func TestHandleWebMessageSupportsPayloadStateChange(t *testing.T) {
	e := &Chromium{}
	handled := e.HandleWebMessage(`{"type":"stateChange","payload":{"loading":true,"url":"https://example.com","title":"Example"}}`)

	if !handled {
		t.Fatalf("expected message to be handled")
	}
	if !e.state.isLoading {
		t.Fatalf("expected loading state to be true")
	}
	if e.state.currentURL != "https://example.com" {
		t.Fatalf("expected url to be updated, got %q", e.state.currentURL)
	}
	if e.state.currentTitle != "Example" {
		t.Fatalf("expected title to be updated, got %q", e.state.currentTitle)
	}
}

func TestHandleWebMessageSupportsLegacyStateChange(t *testing.T) {
	e := &Chromium{}
	handled := e.HandleWebMessage(`{"type":"stateChange","loading":false,"url":"https://legacy.example.com","title":"Legacy"}`)

	if !handled {
		t.Fatalf("expected legacy message to be handled")
	}
	if e.state.isLoading {
		t.Fatalf("expected loading state to be false")
	}
	if e.state.currentURL != "https://legacy.example.com" {
		t.Fatalf("expected legacy url to be updated, got %q", e.state.currentURL)
	}
	if e.state.currentTitle != "Legacy" {
		t.Fatalf("expected legacy title to be updated, got %q", e.state.currentTitle)
	}
}

func TestHandleDispatchMessageRunsQueuedWork(t *testing.T) {
	e := &Chromium{}
	ran := false

	e.dispatch.queue = append(e.dispatch.queue, func() {
		ran = true
	})

	if !e.HandleDispatchMessage("__dispatch__") {
		t.Fatalf("expected dispatch message to be consumed")
	}
	if !ran {
		t.Fatalf("expected queued work to run")
	}
}

func TestHandleWebMessageRejectsUnknownMessage(t *testing.T) {
	e := &Chromium{}
	if e.HandleWebMessage(`{"foo":"bar"}`) {
		t.Fatalf("expected unknown message to be ignored")
	}
}
