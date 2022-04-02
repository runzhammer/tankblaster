package engine

type SpriteSpec struct {
	Still struct {
		Speed  float64 `yaml:"speed"`
		Frames []FrameSpec
	} `yaml:"still"`
	Stand struct {
		Speed  float64 `yaml:"speed"`
		Frames []FrameSpec
	} `yaml:"stand"`
	Drive struct {
		Speed  float64 `yaml:"speed"`
		Frames []FrameSpec
	} `yaml:"drive"`
}

type FrameSpec struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
	W int `yaml:"w"`
	H int `yaml:"h"`
}
