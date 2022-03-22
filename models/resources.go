package models

import _ "embed"

//go:embed resources/tank.png
var tankSprite []byte

//go:embed resources/tank.yaml
var tankSpec []byte

//go:embed resources/background.png
var backgroundSprite []byte

//go:embed resources/background.yaml
var backgroundSpec []byte

//go:embed resources/ground.png
var groundSprite []byte

//go:embed resources/ground.yaml
var groundSpec []byte
