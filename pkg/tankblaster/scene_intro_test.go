package tankblaster

import (
	"io"
	"testing"

	"codeberg.org/rabenauge/soundsetgo"
	"github.com/hajimehoshi/ebiten/v2"
	r "github.com/runzhammer/tankblaster/resources"
)

func TestIntroAssetsDecode(t *testing.T) {
	paths := []string{
		"intro/Intro_II_v2.bmp",
		"intro/Intro_Text_1.png",
		"intro/Intro_Text_2.png",
		"intro/Intro_Text_3.png",
		"intro/Intro_Text_4.png",
		"intro/Intro_Text_5.png",
		"intro/Intro_Text_6.png",
	}
	for _, path := range paths {
		data, err := r.IntroBytes(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		img, err := decodeIntroImage(path, data)
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		if img.Bounds().Empty() {
			t.Fatalf("%s decoded to empty image", path)
		}
	}
}

func TestIntroMODRendersPCM(t *testing.T) {
	for _, path := range []string{"intro/TECHNOMN.MOD", hallOfFameMusicPath} {
		data, err := r.IntroBytes(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		snd, err := soundsetgo.Decode(path, data, soundsetgo.Options{
			SampleRate:      AudioSampleRate,
			DurationSeconds: 1,
		})
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		buf := make([]int16, 4096)
		n, err := snd.RenderPCM(buf)
		if err != nil && err != io.EOF {
			t.Fatalf("render %s: %v", path, err)
		}
		if n == 0 {
			t.Fatalf("%s rendered no samples", path)
		}
	}
}

func TestIntroTimeline(t *testing.T) {
	s := &introScene{texts: make([]*ebiten.Image, 6)}
	if s.colorStartFrame() != 3180 {
		t.Fatalf("colorStartFrame = %d, want 3180", s.colorStartFrame())
	}
	if s.endFrame() != 3570 {
		t.Fatalf("endFrame = %d, want 3570", s.endFrame())
	}
	if _, _, ok := s.textState(); ok {
		t.Fatal("text visible before initial delay")
	}
	s.tick = introInitialDelayFrames + introRevealFrames
	index, local, ok := s.textState()
	if !ok || index != 0 || local != introRevealFrames {
		t.Fatalf("text state = index %d local %d ok %t, want first text held", index, local, ok)
	}
}
