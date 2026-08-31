package tankblaster

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	weaponspkg "github.com/runzhammer/gamedemo/pkg/tankblaster/weapons"
	r "github.com/runzhammer/gamedemo/resources"
)

type soundPlayer struct {
	context *audio.Context
	cache   map[string][]byte
	loops   map[string]*audio.Player
	repeats map[string]chan struct{}
	mu      sync.Mutex
}

func newSoundPlayer() *soundPlayer {
	return &soundPlayer{
		context: audio.NewContext(AudioSampleRate),
		cache:   make(map[string][]byte),
		loops:   make(map[string]*audio.Player),
		repeats: make(map[string]chan struct{}),
	}
}

func (p *soundPlayer) Play(path string) {
	p.PlayWithOptions(path, soundOptions{})
}

func (p *soundPlayer) PlayWithOptions(path string, opts soundOptions) {
	path = strings.TrimSpace(path)
	if path == "" || p == nil || p.context == nil {
		return
	}
	path = selectedSoundPath(path, opts)
	opts = normalizedSoundOptions(opts)
	if opts.Loop {
		p.EnsureLoopWithOptions(path, path, opts)
		return
	}
	if opts.RepeatEverySeconds > 0 {
		p.playRepeated(path, opts)
		return
	}
	samples, err := p.samples(path, opts.PlaybackSpeed)
	if err != nil {
		log.Printf("sound %q: %v", path, err)
		return
	}
	samples = trimmedSamples(samples, opts)
	if len(samples) == 0 {
		return
	}
	player := audio.NewPlayerFromBytes(p.context, samples)
	player.SetVolume(opts.Volume)
	player.Play()
}

func (p *soundPlayer) EnsureLoop(key, path string) {
	p.EnsureLoopWithOptions(key, path, soundOptions{Loop: true})
}

func (p *soundPlayer) EnsureLoopWithOptions(key, path string, opts soundOptions) {
	key = strings.TrimSpace(key)
	path = strings.TrimSpace(path)
	if key == "" || path == "" || p == nil || p.context == nil {
		return
	}
	path = selectedSoundPath(path, opts)
	opts = normalizedSoundOptions(opts)
	if !opts.KeepSilenceForLoop {
		opts.TrimSilenceForLoop = true
	}
	if opts.Loop && opts.RepeatEverySeconds > 0 {
		p.ensureRepeatedLoop(key, path, opts)
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

	samples, err := p.samples(path, opts.PlaybackSpeed)
	if err != nil {
		log.Printf("sound loop %q: %v", path, err)
		return
	}
	samples = trimmedSamples(samples, soundOptions{StartOffsetSeconds: opts.StartOffsetSeconds})
	if opts.TrimSilenceForLoop {
		samples = trimSilentSampleEdges(samples, opts.SilenceThreshold)
	}
	if len(samples) == 0 {
		return
	}
	player, err := audio.NewPlayer(p.context, audio.NewInfiniteLoop(bytes.NewReader(samples), int64(len(samples))))
	if err != nil {
		log.Printf("sound loop %q: %v", path, err)
		return
	}
	player.SetVolume(opts.Volume)

	p.mu.Lock()
	if existing := p.loops[key]; existing != nil {
		p.mu.Unlock()
		_ = player.Close()
		return
	}
	p.loops[key] = player
	p.mu.Unlock()
	player.Play()
	if opts.DurationSeconds > 0 {
		go func() {
			time.Sleep(soundDuration(opts.DurationSeconds))
			p.StopLoop(key)
		}()
	}
}

func (p *soundPlayer) ensureRepeatedLoop(key, path string, opts soundOptions) {
	p.mu.Lock()
	if stop := p.repeats[key]; stop != nil {
		p.mu.Unlock()
		return
	}
	stop := make(chan struct{})
	p.repeats[key] = stop
	p.mu.Unlock()

	go func() {
		defer func() {
			p.mu.Lock()
			if p.repeats[key] == stop {
				delete(p.repeats, key)
			}
			p.mu.Unlock()
		}()
		interval := soundDuration(opts.RepeatEverySeconds)
		if interval <= 0 {
			return
		}
		start := time.Now()
		for repeats := 0; ; repeats++ {
			if opts.MaxRepeats > 0 && repeats >= opts.MaxRepeats {
				return
			}
			if opts.DurationSeconds > 0 && time.Since(start) > soundDuration(opts.DurationSeconds) {
				return
			}
			single := opts
			single.Loop = false
			single.RepeatEverySeconds = 0
			single.MaxRepeats = 1
			p.PlayWithOptions(path, single)
			select {
			case <-stop:
				return
			case <-time.After(interval):
			}
		}
	}()
}

func (p *soundPlayer) StopLoop(key string) {
	key = strings.TrimSpace(key)
	if key == "" || p == nil {
		return
	}
	p.mu.Lock()
	player := p.loops[key]
	delete(p.loops, key)
	stop := p.repeats[key]
	delete(p.repeats, key)
	p.mu.Unlock()
	if player != nil {
		player.Pause()
		_ = player.Close()
	}
	if stop != nil {
		close(stop)
	}
}

func (p *soundPlayer) StopAllLoops() {
	if p == nil {
		return
	}
	p.mu.Lock()
	loops := p.loops
	p.loops = make(map[string]*audio.Player)
	repeats := p.repeats
	p.repeats = make(map[string]chan struct{})
	p.mu.Unlock()
	for _, player := range loops {
		if player != nil {
			player.Pause()
			_ = player.Close()
		}
	}
	for _, stop := range repeats {
		if stop != nil {
			close(stop)
		}
	}
}

func (p *soundPlayer) playRepeated(path string, opts soundOptions) {
	repeats := opts.MaxRepeats
	if repeats <= 0 {
		repeats = 1
	}
	interval := soundDuration(opts.RepeatEverySeconds)
	single := opts
	single.RepeatEverySeconds = 0
	single.MaxRepeats = 1
	go func() {
		start := time.Now()
		for i := 0; i < repeats; i++ {
			if opts.DurationSeconds > 0 && time.Since(start) > soundDuration(opts.DurationSeconds) {
				return
			}
			p.PlayWithOptions(path, single)
			if i == repeats-1 {
				return
			}
			time.Sleep(interval)
		}
	}()
}

func (p *soundPlayer) samples(path string, speed float64) ([]byte, error) {
	speed = normalizedPlaybackSpeed(speed)
	cacheKey := path
	if speed != 1 {
		cacheKey = path + "@speed=" + strconvFormatFloat(speed)
	}
	p.mu.Lock()
	if samples, ok := p.cache[cacheKey]; ok {
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
	if speed != 1 {
		resampled := audio.Resample(bytes.NewReader(samples), int64(len(samples)), int(float64(AudioSampleRate)*speed), AudioSampleRate)
		samples, err = io.ReadAll(resampled)
		if err != nil {
			return nil, err
		}
	}

	p.mu.Lock()
	p.cache[cacheKey] = samples
	p.mu.Unlock()
	return samples, nil
}

func (g *GameLoop) playSound(path string) {
	g.playSoundWithOptions(path, configuredSoundOptions(path))
}

func (g *GameLoop) playSoundWithOptions(path string, opts soundOptions) {
	if g == nil || g.muted || strings.TrimSpace(path) == "" {
		return
	}
	if g.sounds == nil {
		g.sounds = newSoundPlayer()
	}
	g.sounds.PlayWithOptions(path, opts)
}

func (g *GameLoop) playSoundLoop(key, path string) {
	opts := configuredSoundOptions(path)
	opts.Loop = true
	g.playSoundLoopWithOptions(key, path, opts)
}

func (g *GameLoop) playSoundLoopWithOptions(key, path string, opts soundOptions) {
	if g == nil || g.muted || strings.TrimSpace(path) == "" {
		return
	}
	if g.sounds == nil {
		g.sounds = newSoundPlayer()
	}
	g.sounds.EnsureLoopWithOptions(key, path, opts)
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
	path := tankBlasterSounds.Events[event]
	opts := configuredSoundOptions(path)
	opts.Loop = true
	s.g.playSoundLoopWithOptions(key, path, opts)
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
	s.playSound(s.weaponImpactSoundPath(weapon))
}

func (s *GameScene) weaponImpactSoundPath(weapon weaponspkg.Weapon) string {
	return selectedSoundPath(weapon.ImpactSound, configuredSoundOptions(weapon.ImpactSound))
}

func configuredSoundOptions(path string) soundOptions {
	path = strings.TrimSpace(path)
	if path == "" {
		return normalizedSoundOptions(soundOptions{})
	}
	opts := tankBlasterSounds.Options[path]
	return normalizedSoundOptions(opts)
}

func normalizedSoundOptions(opts soundOptions) soundOptions {
	opts.PlaybackSpeed = normalizedPlaybackSpeed(opts.PlaybackSpeed)
	if opts.Volume <= 0 {
		opts.Volume = 1
	}
	if opts.DurationSeconds < 0 {
		opts.DurationSeconds = 0
	}
	if opts.RepeatEverySeconds < 0 {
		opts.RepeatEverySeconds = 0
	}
	if opts.StartOffsetSeconds < 0 {
		opts.StartOffsetSeconds = 0
	}
	if opts.SilenceThreshold <= 0 {
		opts.SilenceThreshold = 96
	}
	return opts
}

func normalizedPlaybackSpeed(speed float64) float64 {
	if speed <= 0 || math.IsNaN(speed) || math.IsInf(speed, 0) {
		return 1
	}
	return math.Max(0.1, math.Min(4, speed))
}

func selectedSoundPath(path string, opts soundOptions) string {
	if len(opts.AlternatePaths) == 0 {
		return strings.TrimSpace(path)
	}
	paths := make([]string, 0, len(opts.AlternatePaths)+1)
	if path = strings.TrimSpace(path); path != "" {
		paths = append(paths, path)
	}
	for _, candidate := range opts.AlternatePaths {
		if candidate = strings.TrimSpace(candidate); candidate != "" {
			paths = append(paths, candidate)
		}
	}
	if len(paths) == 0 {
		return ""
	}
	return paths[rand.Intn(len(paths))]
}

func trimmedSamples(samples []byte, opts soundOptions) []byte {
	if len(samples) == 0 {
		return nil
	}
	start := bytesForSeconds(opts.StartOffsetSeconds)
	if start >= len(samples) {
		return nil
	}
	if start > 0 {
		samples = samples[start:]
	}
	if opts.DurationSeconds <= 0 {
		return samples
	}
	limit := bytesForSeconds(opts.DurationSeconds)
	if limit <= 0 {
		return nil
	}
	if limit >= len(samples) {
		return samples
	}
	return samples[:limit]
}

func trimSilentSampleEdges(samples []byte, threshold int16) []byte {
	if len(samples) < 4 {
		return samples
	}
	frameCount := len(samples) / 4
	startFrame := 0
	for startFrame < frameCount && stereoFrameSilent(samples[startFrame*4:startFrame*4+4], threshold) {
		startFrame++
	}
	endFrame := frameCount - 1
	for endFrame >= startFrame && stereoFrameSilent(samples[endFrame*4:endFrame*4+4], threshold) {
		endFrame--
	}
	if startFrame == 0 && endFrame == frameCount-1 {
		return samples[:frameCount*4]
	}
	if endFrame < startFrame {
		return samples[:0]
	}
	return samples[startFrame*4 : (endFrame+1)*4]
}

func stereoFrameSilent(frame []byte, threshold int16) bool {
	if len(frame) < 4 {
		return true
	}
	left := int16(uint16(frame[0]) | uint16(frame[1])<<8)
	right := int16(uint16(frame[2]) | uint16(frame[3])<<8)
	limit := int(threshold)
	return sampleAbsInt(int(left)) <= limit && sampleAbsInt(int(right)) <= limit
}

func sampleAbsInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func bytesForSeconds(seconds float64) int {
	if seconds <= 0 {
		return 0
	}
	bytesPerSecond := AudioSampleRate * 2 * 2
	n := int(math.Round(seconds * float64(bytesPerSecond)))
	return n - n%4
}

func soundDuration(seconds float64) time.Duration {
	return time.Duration(math.Round(seconds * float64(time.Second)))
}

func strconvFormatFloat(value float64) string {
	return fmt.Sprintf("%.3f", value)
}
