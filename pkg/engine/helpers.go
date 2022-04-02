package engine

import (
	"bytes"
	"image"
	"log"

	"gopkg.in/yaml.v2"
)

type Number interface {
	int64 | float64 | int | uint64
}

func MakeRange[T Number](min, max T) []T {
	a := make([]T, int(max-min+1))
	for i := range a {
		a[i] = min + T(i)
	}
	return a
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
