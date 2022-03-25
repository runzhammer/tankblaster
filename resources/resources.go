package resources

import _ "embed"

//go:embed tank.png
var tankSprite []byte

//go:embed tank.yaml
var tankSpec []byte

//go:embed background.png
var backgroundSprite []byte

//go:embed background.yaml
var backgroundSpec []byte

//go:embed ground.png
var groundSprite []byte

//go:embed ground.yaml
var groundSpec []byte
