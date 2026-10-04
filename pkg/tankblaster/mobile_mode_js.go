//go:build js

package tankblaster

import (
	"strings"
	"syscall/js"
)

func mobileControlsEnabled() bool {
	search := js.Global().Get("location").Get("search").String()
	if strings.Contains(search, "desktop=1") {
		return false
	}
	if strings.Contains(search, "mobile=1") {
		return true
	}

	nav := js.Global().Get("navigator")
	maxTouchPoints := nav.Get("maxTouchPoints").Int()
	ua := strings.ToLower(nav.Get("userAgent").String())
	mobileUA := strings.Contains(ua, "android") ||
		strings.Contains(ua, "iphone") ||
		strings.Contains(ua, "ipad") ||
		strings.Contains(ua, "ipod") ||
		strings.Contains(ua, "mobile")
	coarsePointer := false
	if matchMedia := js.Global().Get("matchMedia"); matchMedia.Type() == js.TypeFunction {
		coarsePointer = matchMedia.Invoke("(pointer: coarse)").Get("matches").Bool()
	}
	return maxTouchPoints > 0 && (mobileUA || coarsePointer)
}
