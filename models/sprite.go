package models

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"gopkg.in/yaml.v3"
)

type Sprite struct {
	Image      *ebiten.Image
	Animations Animations
}

type Animations struct {
	Stand struct {
		Speed  float64 `yaml:"speed"`
		Frames []frameSpec
	} `yaml:"stand"`
	Move struct {
		Speed  float64 `yaml:"speed"`
		Frames []frameSpec
	} `yaml:"move"`
}

type frameSpec struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
	W int `yaml:"w"`
	H int `yaml:"h"`
}

func (t *Sprite) Init(sprite []byte, animations []byte) {

	var err error

	// load sprite
	spriteDecoded, _, err := image.Decode(bytes.NewReader(sprite))
	if err != nil {
		log.Fatal(err)
	}

	t.Image = ebiten.NewImageFromImage(spriteDecoded)

	if err != nil {
		log.Fatal(err)
	}

	// load frames
	err = yaml.Unmarshal(animations, &t.Animations)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}
