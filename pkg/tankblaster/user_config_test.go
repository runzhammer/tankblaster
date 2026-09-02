package tankblaster

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserConfigRoundTrip(t *testing.T) {
	originalLanguage := currentLanguage
	t.Cleanup(func() {
		currentLanguage = originalLanguage
	})

	dir := t.TempDir()
	path := filepath.Join(dir, userConfigFileName)
	currentLanguage = languageEnglish
	game := &GameLoop{
		rounds: 42,
		options: gameOptions{
			projectileReentry: 2,
			palmCount:         1,
			cloudAggression:   65,
			quickRoundStart:   false,
		},
	}

	if err := game.saveUserConfigToPath(path); err != nil {
		t.Fatalf("save user config: %v", err)
	}

	currentLanguage = languageGerman
	loaded := &GameLoop{
		rounds:  10,
		options: defaultGameOptions(),
	}
	if err := loaded.loadUserConfigFromPath(path); err != nil {
		t.Fatalf("load user config: %v", err)
	}

	if got, want := loaded.rounds, 42; got != want {
		t.Fatalf("rounds = %d, want %d", got, want)
	}
	if got, want := loaded.options.projectileReentry, 2; got != want {
		t.Fatalf("projectileReentry = %d, want %d", got, want)
	}
	if got, want := loaded.options.palmCount, 1; got != want {
		t.Fatalf("palmCount = %d, want %d", got, want)
	}
	if got, want := loaded.options.cloudAggression, 65; got != want {
		t.Fatalf("cloudAggression = %d, want %d", got, want)
	}
	if got, want := loaded.options.quickRoundStart, false; got != want {
		t.Fatalf("quickRoundStart = %t, want %t", got, want)
	}
	if got, want := currentLanguage, languageEnglish; got != want {
		t.Fatalf("currentLanguage = %d, want %d", got, want)
	}
}

func TestSaveUserConfigPreservesOnlineIdentity(t *testing.T) {
	path := filepath.Join(t.TempDir(), userConfigFileName)
	onlineRounds := 8
	id := &onlineIdentity{
		PlayerID:    "player-1",
		PlayerToken: "secret",
		DisplayName: "Nina",
	}
	if err := writeUserConfigFile(path, userConfigFile{OnlineRounds: &onlineRounds, OnlineIdentity: id}); err != nil {
		t.Fatalf("write config: %v", err)
	}

	game := &GameLoop{
		rounds:  7,
		options: defaultGameOptions(),
	}
	if err := game.saveUserConfigToPath(path); err != nil {
		t.Fatalf("save user config: %v", err)
	}

	cfg, err := readUserConfigFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if cfg.OnlineIdentity == nil {
		t.Fatal("online identity missing")
	}
	if got, want := cfg.OnlineIdentity.PlayerID, id.PlayerID; got != want {
		t.Fatalf("player id = %q, want %q", got, want)
	}
	if got, want := cfg.OnlineIdentity.PlayerToken, id.PlayerToken; got != want {
		t.Fatalf("player token = %q, want %q", got, want)
	}
	if got, want := cfg.OnlineIdentity.DisplayName, id.DisplayName; got != want {
		t.Fatalf("display name = %q, want %q", got, want)
	}
	if cfg.OnlineRounds == nil || *cfg.OnlineRounds != onlineRounds {
		t.Fatalf("online rounds = %v, want %d", cfg.OnlineRounds, onlineRounds)
	}
}

func TestSaveOnlineIdentityPreservesUserConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), userConfigFileName)
	rounds := 12
	if err := writeUserConfigFile(path, userConfigFile{Rounds: &rounds}); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := saveOnlineIdentityToPath(path, onlineIdentity{
		PlayerID:    "player-2",
		PlayerToken: "token",
		DisplayName: "Alex",
	}); err != nil {
		t.Fatalf("save online identity: %v", err)
	}

	cfg, err := readUserConfigFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if cfg.Rounds == nil || *cfg.Rounds != rounds {
		t.Fatalf("rounds = %v, want %d", cfg.Rounds, rounds)
	}
	if cfg.OnlineIdentity == nil {
		t.Fatal("online identity missing")
	}
	if got, want := cfg.OnlineIdentity.DisplayName, "Alex"; got != want {
		t.Fatalf("display name = %q, want %q", got, want)
	}
}

func TestLoadOnlineIdentityFromUserConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), userConfigFileName)
	if err := saveOnlineIdentityToPath(path, onlineIdentity{
		PlayerID:    "player-3",
		PlayerToken: "token-3",
		DisplayName: "Sam",
	}); err != nil {
		t.Fatalf("save online identity: %v", err)
	}

	id := loadOnlineIdentityFromPath(path)
	if got, want := id.PlayerID, "player-3"; got != want {
		t.Fatalf("player id = %q, want %q", got, want)
	}
	if got, want := id.PlayerToken, "token-3"; got != want {
		t.Fatalf("player token = %q, want %q", got, want)
	}
	if got, want := id.DisplayName, "Sam"; got != want {
		t.Fatalf("display name = %q, want %q", got, want)
	}
}

func TestSaveOnlineRoundsPreservesUserConfigAndIdentity(t *testing.T) {
	path := filepath.Join(t.TempDir(), userConfigFileName)
	rounds := 12
	id := &onlineIdentity{
		PlayerID:    "player-4",
		PlayerToken: "token-4",
		DisplayName: "Mika",
	}
	if err := writeUserConfigFile(path, userConfigFile{Rounds: &rounds, OnlineIdentity: id}); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := saveOnlineRoundsToPath(path, 17); err != nil {
		t.Fatalf("save online rounds: %v", err)
	}

	cfg, err := readUserConfigFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if cfg.Rounds == nil || *cfg.Rounds != rounds {
		t.Fatalf("rounds = %v, want %d", cfg.Rounds, rounds)
	}
	if cfg.OnlineRounds == nil || *cfg.OnlineRounds != 17 {
		t.Fatalf("online rounds = %v, want %d", cfg.OnlineRounds, 17)
	}
	if cfg.OnlineIdentity == nil || cfg.OnlineIdentity.PlayerID != id.PlayerID {
		t.Fatalf("online identity = %#v, want player id %q", cfg.OnlineIdentity, id.PlayerID)
	}
}

func TestLoadOnlineRoundsFromUserConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), userConfigFileName)
	if err := saveOnlineRoundsToPath(path, 23); err != nil {
		t.Fatalf("save online rounds: %v", err)
	}

	if got, want := loadOnlineRoundsFromPath(path, 5), 23; got != want {
		t.Fatalf("online rounds = %d, want %d", got, want)
	}
}

func TestSaveUserConfigCreatesParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", userConfigFileName)
	game := &GameLoop{
		rounds:  10,
		options: defaultGameOptions(),
	}

	if err := game.saveUserConfigToPath(path); err != nil {
		t.Fatalf("save user config: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat saved user config: %v", err)
	}
}

func TestLoadUserConfigKeepsDefaultsForMissingFile(t *testing.T) {
	game := &GameLoop{
		rounds: 10,
		options: gameOptions{
			projectileReentry: 1,
			palmCount:         -1,
			cloudAggression:   10,
			quickRoundStart:   true,
		},
	}

	if err := game.loadUserConfigFromPath(filepath.Join(t.TempDir(), userConfigFileName)); err != nil {
		t.Fatalf("load missing user config: %v", err)
	}

	if got, want := game.rounds, 10; got != want {
		t.Fatalf("rounds = %d, want %d", got, want)
	}
	if got, want := game.options.cloudAggression, 10; got != want {
		t.Fatalf("cloudAggression = %d, want %d", got, want)
	}
}

func TestLoadUserConfigClampsInvalidValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), userConfigFileName)
	data := []byte("rounds: 250\noptions:\n  projectile_reentry: 99\n  palm_count: -5\n  cloud_aggression: 150\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	game := &GameLoop{
		rounds:  10,
		options: defaultGameOptions(),
	}
	if err := game.loadUserConfigFromPath(path); err != nil {
		t.Fatalf("load user config: %v", err)
	}

	if got, want := game.rounds, 99; got != want {
		t.Fatalf("rounds = %d, want %d", got, want)
	}
	if got, want := game.options.projectileReentry, 1; got != want {
		t.Fatalf("projectileReentry = %d, want %d", got, want)
	}
	if got, want := game.options.palmCount, -1; got != want {
		t.Fatalf("palmCount = %d, want %d", got, want)
	}
	if got, want := game.options.cloudAggression, 100; got != want {
		t.Fatalf("cloudAggression = %d, want %d", got, want)
	}
}
