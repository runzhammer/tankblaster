//go:build !js

package tankblaster

import (
	"bytes"
	"encoding/binary"

	"codeberg.org/rabenauge/soundsetgo"
	_ "codeberg.org/rabenauge/soundsetgo/formats/mod"
)

func modSamples(path string, data []byte) ([]byte, error) {
	snd, err := soundsetgo.Decode(path, data, soundsetgo.Options{
		SampleRate:      AudioSampleRate,
		DurationSeconds: 60,
	})
	if err != nil {
		return nil, err
	}
	pcm, err := soundsetgo.RenderAll(snd)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, pcm); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
