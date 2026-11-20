package ui

import (
	"testing"
	"time"
)

func TestRawKeyConstants(t *testing.T) {
	if KeyEnter == KeyEscape {
		t.Fatal("KeyEnter and KeyEscape should be distinct")
	}
	if KeyCtrlC == KeyNone {
		t.Fatal("KeyCtrlC should not be KeyNone")
	}
}

func TestDoneChannelListenKey(t *testing.T) {
	stopChan := make(chan RawKey, 1)
	done := make(chan struct{})

	// Closing done channel should cause goroutine to exit without blocking
	close(done)
	ListenKeyForRecording(stopChan, done)

	select {
	case <-stopChan:
		t.Fatal("stopChan should not receive when done was closed immediately")
	case <-time.After(50 * time.Millisecond):
		// Exited promptly
	}
}
