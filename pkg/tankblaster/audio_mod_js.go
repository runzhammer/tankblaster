//go:build js

package tankblaster

import "fmt"

func modSamples(path string, _ []byte) ([]byte, error) {
	return nil, fmt.Errorf("MOD playback is not available in WebAssembly: %s", path)
}
