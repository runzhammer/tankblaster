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
	loops   map[string]*audio.Player
	mu      sync.Mutex
}

func newSoundPlayer() *soundPlayer {
	return &soundPlayer{
		context: audio.NewContext(AudioSampleRate),
		cache:   make(map[string][]byte),
		loops:   make(map[string]*audio.Player),
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

func (p *soundPlayer) EnsureLoop(key, path string) {
	key = strings.TrimSpace(key)
	path = strings.TrimSpace(path)
	if key == "" || path == "" || p == nil || p.context == nil {
		return
	}
	p.mu.Lock()
	if player := p.loops[key]; player != nil {
		if !player.IsPlaying() {
			_ = player.Rewind()
			player.Play()
		}
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	samples, err := p.samples(path)
	if err != nil {
		log.Printf("sound loop %q: %v", path, err)
		return
	}
	player, err := audio.NewPlayer(p.context, audio.NewInfiniteLoop(bytes.NewReader(samples), int64(len(samples))))
	if err != nil {
		log.Printf("sound loop %q: %v", path, err)
		return
	}

	p.mu.Lock()
	if existing := p.loops[key]; existing != nil {
		p.mu.Unlock()
		_ = player.Close()
		return
	}
	p.loops[key] = player
	p.mu.Unlock()
	player.Play()
}

func (p *soundPlayer) StopLoop(key string) {
	key = strings.TrimSpace(key)
	if key == "" || p == nil {
		return
	}
	p.mu.Lock()
	player := p.loops[key]
	delete(p.loops, key)
	p.mu.Unlock()
	if player != nil {
		player.Pause()
		_ = player.Close()
	}
}

func (p *soundPlayer) StopAllLoops() {
	if p == nil {
		return
	}
	p.mu.Lock()
	loops := p.loops
	p.loops = make(map[string]*audio.Player)
	p.mu.Unlock()
	for _, player := range loops {
		if player != nil {
			player.Pause()
			_ = player.Close()
		}
	}
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

func (g *GameLoop) playSoundLoop(key, path string) {
	if g == nil || g.muted || strings.TrimSpace(path) == "" {
		return
	}
	if g.sounds == nil {
		g.sounds = newSoundPlayer()
	}
	g.sounds.EnsureLoop(key, path)
}

func (g *GameLoop) stopSoundLoop(key string) {
	if g == nil || g.sounds == nil {
		return
	}
	g.sounds.StopLoop(key)
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

func (s *GameScene) playEventSoundLoop(key string, event soundEvent) {
	if s == nil || s.g == nil {
		return
	}
	s.g.playSoundLoop(key, tankBlasterSounds.Events[event])
}

func (s *GameScene) stopSoundLoop(key string) {
	if s == nil || s.g == nil {
		return
	}
	s.g.stopSoundLoop(key)
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
