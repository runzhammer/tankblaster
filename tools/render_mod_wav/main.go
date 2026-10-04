package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"codeberg.org/rabenauge/soundsetgo"
	_ "codeberg.org/rabenauge/soundsetgo/formats/mod"
)

const (
	sampleRate      = 44100
	channels        = 2
	bitsPerSample   = 16
	durationSeconds = 60
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s input.mod output.wav\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}
	if err := render(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func render(inPath, outPath string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}
	snd, err := soundsetgo.Decode(inPath, data, soundsetgo.Options{
		SampleRate:      sampleRate,
		DurationSeconds: durationSeconds,
	})
	if err != nil {
		return err
	}
	pcm, err := soundsetgo.RenderAll(snd)
	if err != nil {
		return err
	}

	var audioData bytes.Buffer
	if err := binary.Write(&audioData, binary.LittleEndian, pcm); err != nil {
		return err
	}
	return os.WriteFile(outPath, wavBytes(audioData.Bytes()), 0644)
}

func wavBytes(audioData []byte) []byte {
	byteRate := sampleRate * channels * bitsPerSample / 8
	blockAlign := channels * bitsPerSample / 8

	var out bytes.Buffer
	out.WriteString("RIFF")
	writeU32(&out, uint32(36+len(audioData)))
	out.WriteString("WAVE")
	out.WriteString("fmt ")
	writeU32(&out, 16)
	writeU16(&out, 1)
	writeU16(&out, channels)
	writeU32(&out, sampleRate)
	writeU32(&out, uint32(byteRate))
	writeU16(&out, uint16(blockAlign))
	writeU16(&out, bitsPerSample)
	out.WriteString("data")
	writeU32(&out, uint32(len(audioData)))
	out.Write(audioData)
	return out.Bytes()
}

func writeU16(buf *bytes.Buffer, v uint16) {
	_ = binary.Write(buf, binary.LittleEndian, v)
}

func writeU32(buf *bytes.Buffer, v uint32) {
	_ = binary.Write(buf, binary.LittleEndian, v)
}
