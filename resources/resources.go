package resources

import _ "embed"

//go:embed config.yaml
var GameConfig []byte

//go:embed tank.png
var TankSprite []byte

//go:embed tank.yaml
var TankSpec []byte

//go:embed cannon.png
var CannonSprite []byte

//go:embed cannon.yaml
var CannonSpec []byte

//go:embed background.png
var BackgroundSprite []byte

//go:embed background.yaml
var BackgroundSpec []byte

//go:embed ground.png
var GroundSprite []byte
