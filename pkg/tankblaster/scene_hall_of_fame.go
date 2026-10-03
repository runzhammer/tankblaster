package tankblaster

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/runzhammer/tankblaster/pkg/core"
	r "github.com/runzhammer/tankblaster/resources"
)

var _ core.Scene = (*hallOfFameScene)(nil)

const (
	hallOfFameMusicLoopKey = "hall_of_fame_music"
	hallOfFameMusicPath    = "intro/RIFF_RFF.MOD"
)

type hallOfFameScore struct {
	Name  string
	Score int
	Color color.RGBA
	Index int
}

type hallOfFameScene struct {
	g          *GameLoop
	background *ebiten.Image
	title      *ebiten.Image
	cloud      *ebiten.Image
	cloudFace  *ebiten.Image
	scores     []hallOfFameScore
	rng        *rand.Rand
	cloudPos   hallOfFameVec
	cloudVel   hallOfFameVec
	musicLive  bool
}

type hallOfFameVec struct {
	X float64
	Y float64
}

func NewHallOfFameScene(scores []hallOfFameScore) func(*GameLoop) (core.Scene, error) {
	copied := make([]hallOfFameScore, len(scores))
	copy(copied, scores)
	return func(game *GameLoop) (core.Scene, error) {
		return newHallOfFameScene(game, copied)
	}
}

func newHallOfFameScene(game *GameLoop, scores []hallOfFameScore) (core.Scene, error) {
	scene := &hallOfFameScene{
		g:          game,
		background: mustImageFromPNG(r.HallOfFameBackground),
		title:      mustImageFromPNG(r.HallOfFameTitle),
		cloud:      mustImageFromPNG(r.CloudLightning),
		cloudFace:  mustImageFromPNG(r.CloudGrinPNG),
		scores:     sortedHallOfFameScores(scores),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
		cloudPos:   hallOfFameVec{X: 140, Y: 170},
		cloudVel:   hallOfFameVec{X: 2.6, Y: 1.2},
	}
	return scene, nil
}

func NewDebugHallOfFameScene(game *GameLoop) (core.Scene, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	names := []string{"Spieler 1", "Spieler 2", "F. Merz", "A. Merkel", "Spongebob"}
	scores := make([]hallOfFameScore, 0, len(names))
	for i, name := range names {
		scores = append(scores, hallOfFameScore{
			Name:  name,
			Score: rng.Intn(31) - 8,
			Color: defaultTankColor(i),
			Index: i,
		})
	}
	return newHallOfFameScene(game, scores)
}

func (s *hallOfFameScene) Update() error {
	s.startMusic()
	s.updateCloud()
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		return s.back()
	}
	if primaryPointerJustReleased() {
		x, y := primaryPointerPosition()
		if image.Pt(x, y).In(hallOfFameBackRect()) {
			return s.back()
		}
	}
	return nil
}

func (s *hallOfFameScene) Draw(screen *ebiten.Image) {
	drawScaledImage(screen, s.background, image.Rect(0, 0, ScreenWidth, ScreenHeight))
	s.drawFlyingCloud(screen)
	s.drawTitle(screen)
	s.drawScoreTable(screen)
	drawDialogButton(screen, hallOfFameBackRect(), texts().HallOfFameBack)
}

func (s *hallOfFameScene) startMusic() {
	if s.g == nil || s.musicLive {
		return
	}
	s.musicLive = true
	s.g.playSoundLoopWithOptions(hallOfFameMusicLoopKey, hallOfFameMusicPath, soundOptions{
		Loop:               true,
		Volume:             0.32,
		KeepSilenceForLoop: true,
	})
}

func (s *hallOfFameScene) back() error {
	if s.g != nil {
		s.g.stopSoundLoop(hallOfFameMusicLoopKey)
		return s.g.SetNewScene(NewPlayerSelectionScene)
	}
	return nil
}

func (s *hallOfFameScene) updateCloud() {
	if s.rng == nil {
		return
	}
	s.cloudPos.X += s.cloudVel.X
	s.cloudPos.Y += s.cloudVel.Y
	minX := -70.0
	maxX := float64(ScreenWidth - 100)
	if s.cloudPos.X < minX || s.cloudPos.X > maxX {
		s.cloudVel.X = -s.cloudVel.X
		s.cloudVel.Y = s.rng.Float64()*2.2 - 1.1
		s.cloudPos.X = math.Max(minX, math.Min(maxX, s.cloudPos.X))
	}
	minY := 118.0
	maxY := 500.0
	if s.cloudPos.Y < minY || s.cloudPos.Y > maxY {
		s.cloudVel.Y = -s.cloudVel.Y
		s.cloudVel.X += s.rng.Float64()*0.8 - 0.4
		s.cloudPos.Y = math.Max(minY, math.Min(maxY, s.cloudPos.Y))
	}
	if math.Abs(s.cloudVel.X) < 1.7 {
		s.cloudVel.X = math.Copysign(1.7, s.cloudVel.X)
	}
	if math.Abs(s.cloudVel.X) > 4.6 {
		s.cloudVel.X = math.Copysign(4.6, s.cloudVel.X)
	}
}

func (s *hallOfFameScene) drawFlyingCloud(screen *ebiten.Image) {
	if s.cloud == nil || s.cloudFace == nil {
		return
	}
	cloudRect := image.Rect(int(s.cloudPos.X), int(s.cloudPos.Y), int(s.cloudPos.X)+260, int(s.cloudPos.Y)+117)
	faceRect := image.Rect(cloudRect.Min.X+89, cloudRect.Min.Y+18, cloudRect.Min.X+171, cloudRect.Min.Y+100)
	drawScaledImage(screen, s.cloud, cloudRect)
	drawScaledImage(screen, s.cloudFace, faceRect)
}

func (s *hallOfFameScene) drawTitle(screen *ebiten.Image) {
	if s.title == nil {
		return
	}
	w := 520
	h := w * s.title.Bounds().Dy() / s.title.Bounds().Dx()
	r := image.Rect(ScreenWidth/2-w/2, 36, ScreenWidth/2+w/2, 36+h)
	drawScaledImage(screen, s.title, r)
}

func (s *hallOfFameScene) drawScoreTable(screen *ebiten.Image) {
	t := texts()
	tableW := 620
	rowH := 30
	tableH := 82 + rowH*len(s.scores)
	left := ScreenWidth/2 - tableW/2
	top := 190
	right := left + tableW
	bottom := top + tableH

	drawFilledRect(screen, image.Rect(left-14, top-16, right+14, bottom+12), color.RGBA{R: 22, G: 10, B: 38, A: 118})

	headerY := top + 34
	nameX := left + 80
	scoreX := left + tableW/2 + 60
	statusX := right - 110
	lineColor := color.RGBA{R: 250, G: 246, B: 230, A: 230}
	textColor := color.RGBA{R: 250, G: 246, B: 255, A: 255}

	drawText(screen, t.GameScorePlayer, nameX, headerY, textColor)
	drawText(screen, t.GameScoreSuccess, scoreX, headerY, textColor)
	drawText(screen, t.GameScoreStatus, statusX, headerY, textColor)

	separatorY := headerY + 24
	drawFilledRect(screen, image.Rect(left+18, separatorY, right-18, separatorY+2), lineColor)
	drawFilledRect(screen, image.Rect(left+tableW/2-8, headerY-22, left+tableW/2-6, bottom-18), lineColor)
	drawFilledRect(screen, image.Rect(right-170, headerY-22, right-168, bottom-18), lineColor)

	for i, entry := range s.scores {
		if i >= 8 {
			break
		}
		y := separatorY + 32 + i*rowH
		drawText(screen, strconv.Itoa(i+1)+". "+entry.Name, nameX, y, entry.Color)
		drawText(screen, strconv.Itoa(entry.Score), scoreX+28, y, textColor)
		drawText(screen, "", statusX+18, y, textColor)
	}
}

func hallOfFameBackRect() image.Rectangle {
	return image.Rect(868, 710, 996, 744)
}

func sortedHallOfFameScores(scores []hallOfFameScore) []hallOfFameScore {
	rows := make([]hallOfFameScore, len(scores))
	copy(rows, scores)
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Score != rows[j].Score {
			return rows[i].Score > rows[j].Score
		}
		return rows[i].Index < rows[j].Index
	})
	return rows
}
