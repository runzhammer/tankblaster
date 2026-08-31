//go:build android

package tankblaster

import "sync"

var androidClipboard struct {
	mu      sync.Mutex
	pending string
}

func copyTextToClipboard(value string) error {
	androidClipboard.mu.Lock()
	defer androidClipboard.mu.Unlock()
	androidClipboard.pending = value
	return nil
}

func PendingClipboardText() string {
	androidClipboard.mu.Lock()
	defer androidClipboard.mu.Unlock()
	return androidClipboard.pending
}

func ClearPendingClipboardText(value string) {
	androidClipboard.mu.Lock()
	defer androidClipboard.mu.Unlock()
	if androidClipboard.pending == value {
		androidClipboard.pending = ""
	}
}
