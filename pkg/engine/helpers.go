package engine

import (
	"bytes"
	"image"
	"log"
	"math/rand"
	"time"

	"gopkg.in/yaml.v2"
)

type Number interface {
	int64 | float64 | int | uint64 | float32
}

func MakeRange[T Number](min, max T) []T {
	a := make([]T, int(max-min+1))
	for i := range a {
		a[i] = min + T(i)
	}
	return a
}

func IntRand[T Number](max T) T {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return T(r1.Intn(int(max)))
}

type Resource interface {
	[]byte
}

func LoadImage(imageFile []byte) image.Image {

	image, _, err := image.Decode(bytes.NewReader(imageFile))

	if err != nil {
		log.Fatalf("could not load resource: %s", err)
	}

	return image
}

func LoadFrames(framesFile []byte) SpriteSpec {

	s := SpriteSpec{}

	// load frames
	err := yaml.Unmarshal(framesFile, &s)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return s
}
