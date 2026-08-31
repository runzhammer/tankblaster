//go:build !android

package tankblaster

import (
	"context"
	"sync"

	"golang.design/x/clipboard"
)

var clipboardInit struct {
	once sync.Once
	err  error
}

func copyTextToClipboard(value string) error {
	clipboardInit.once.Do(func() {
		clipboardInit.err = clipboard.Init()
	})
	if clipboardInit.err != nil {
		return clipboardInit.err
	}
	_, err := clipboard.Write(context.Background(), clipboard.FmtText, []byte(value))
	return err
}

func PendingClipboardText() string {
	return ""
}

func ClearPendingClipboardText(value string) {}
