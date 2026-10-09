//go:build js

package tankblaster

import "syscall/js"

func openDefaultBrowser(url string) error {
	js.Global().Get("open").Invoke(url, "_blank", "noopener")
	return nil
}
