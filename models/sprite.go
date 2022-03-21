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
	Animations Animations `yaml:"frames"`
}

type Animations struct {
	Stand struct {
		Tick   float64
		Speed  float64
		Frames []frameSpec
	} `yaml:"stand"`
	Move struct {
		Tick   float64
		Speed  float64
		Frames []frameSpec
	} `yaml:"move"`
}

type frameSpec struct {
	x int `yaml:"x"`
	y int `yaml:"y"`
	w int `yaml:"w"`
	h int `yaml:"h"`
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
	err = yaml.Unmarshal(animations, t.Animations)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}
