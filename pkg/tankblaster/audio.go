package tankblaster

import (
	"bytes"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	weaponspkg "github.com/runzhammer/gamedemo/pkg/tankblaster/weapons"
	r "github.com/runzhammer/gamedemo/resources"
)

type soundPlayer struct {
	context *audio.Context
	cache   map[string][]byte
	mu      sync.Mutex
}

func newSoundPlayer() *soundPlayer {
	return &soundPlayer{
		context: audio.NewContext(AudioSampleRate),
		cache:   make(map[string][]byte),
	}
}

func (p *soundPlayer) Play(path string) {
	path = strings.TrimSpace(path)
	if path == "" || p == nil || p.context == nil {
		return
	}
	samples, err := p.samples(path)
	if err != nil {
		log.Printf("sound %q: %v", path, err)
		return
	}
	audio.NewPlayerFromBytes(p.context, samples).Play()
}

func (p *soundPlayer) samples(path string) ([]byte, error) {
	p.mu.Lock()
	if samples, ok := p.cache[path]; ok {
		p.mu.Unlock()
		return samples, nil
	}
	p.mu.Unlock()

	data, err := r.SoundBytes(path)
	if err != nil {
		return nil, err
	}
	stream, err := wav.Decode(p.context, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	samples, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.cache[path] = samples
	p.mu.Unlock()
	return samples, nil
}

func (g *GameLoop) playSound(path string) {
	if g == nil || g.muted || strings.TrimSpace(path) == "" {
		return
	}
	if g.sounds == nil {
		g.sounds = newSoundPlayer()
	}
	g.sounds.Play(path)
}

func (s *GameScene) playSound(path string) {
	if s == nil || s.g == nil {
		return
	}
	s.g.playSound(path)
}

func (s *GameScene) playEventSound(event soundEvent) {
	s.playSound(tankBlasterSounds.Events[event])
}

func (s *GameScene) playZeroPowerSound(name zeroPowerSound) {
	s.playSound(tankBlasterSounds.ZeroPower[name])
}

func (s *GameScene) playWeaponFireSound(weapon weaponspkg.Weapon) {
	s.playSound(weapon.FireSound)
}

func (s *GameScene) playWeaponImpactSound(weapon weaponspkg.Weapon) {
	s.playSound(weapon.ImpactSound)
}
