package tankblaster

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	userConfigAppDirName = "tankblaster"
	userConfigFileName   = "tank.cfg"
)

type userConfigFile struct {
	Rounds         *int                   `yaml:"rounds,omitempty"`
	OnlineRounds   *int                   `yaml:"online_rounds,omitempty"`
	Options        *userConfigOptionsFile `yaml:"options,omitempty"`
	Language       string                 `yaml:"language,omitempty"`
	OnlineIdentity *onlineIdentity        `yaml:"online_identity,omitempty"`
}

type userConfigOptionsFile struct {
	ProjectileReentry *int  `yaml:"projectile_reentry,omitempty"`
	PalmCount         *int  `yaml:"palm_count,omitempty"`
	CloudAggression   *int  `yaml:"cloud_aggression,omitempty"`
	QuickRoundStart   *bool `yaml:"quick_round_start,omitempty"`
}

var userConfigDirOverride string

func SetUserConfigDir(dir string) {
	userConfigDirOverride = strings.TrimSpace(dir)
}

func (g *GameLoop) loadUserConfig() error {
	path, err := userConfigPath()
	if err != nil {
		return err
	}
	return g.loadUserConfigFromPath(path)
}

func (g *GameLoop) loadUserConfigFromPath(path string) error {
	cfg, err := readUserConfigFile(path)
	if err != nil {
		return err
	}
	g.applyUserConfig(cfg)
	return nil
}

func (g *GameLoop) saveUserConfig() error {
	path, err := userConfigPath()
	if err != nil {
		return err
	}
	return g.saveUserConfigToPath(path)
}

func (g *GameLoop) saveUserConfigToPath(path string) error {
	cfg, err := readUserConfigFile(path)
	if err != nil {
		return err
	}
	cfg.applyGame(g)
	return writeUserConfigFile(path, cfg)
}

func readUserConfigFile(path string) (userConfigFile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return userConfigFile{}, nil
	}
	if err != nil {
		return userConfigFile{}, err
	}

	var cfg userConfigFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return userConfigFile{}, err
	}
	return cfg, nil
}

func writeUserConfigFile(path string, cfg userConfigFile) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (cfg *userConfigFile) applyGame(g *GameLoop) {
	rounds := clampUserRounds(g.rounds)
	reentry := clampProjectileReentry(g.options.projectileReentry)
	palms := clampPalmCount(g.options.palmCount)
	clouds := clampCloudAggression(g.options.cloudAggression)
	quick := g.options.quickRoundStart
	cfg.Rounds = &rounds
	if cfg.Options == nil {
		cfg.Options = &userConfigOptionsFile{}
	}
	cfg.Options.ProjectileReentry = &reentry
	cfg.Options.PalmCount = &palms
	cfg.Options.CloudAggression = &clouds
	cfg.Options.QuickRoundStart = &quick
	cfg.Language = languageConfigValue(currentLanguage)
}

func (g *GameLoop) applyUserConfig(cfg userConfigFile) {
	if cfg.Rounds != nil {
		g.rounds = clampUserRounds(*cfg.Rounds)
	}
	if cfg.Options != nil {
		if cfg.Options.ProjectileReentry != nil {
			g.options.projectileReentry = clampProjectileReentry(*cfg.Options.ProjectileReentry)
		}
		if cfg.Options.PalmCount != nil {
			g.options.palmCount = clampPalmCount(*cfg.Options.PalmCount)
		}
		if cfg.Options.CloudAggression != nil {
			g.options.cloudAggression = clampCloudAggression(*cfg.Options.CloudAggression)
		}
		if cfg.Options.QuickRoundStart != nil {
			g.options.quickRoundStart = *cfg.Options.QuickRoundStart
		}
	}
	if lang, ok := parseLanguageConfigValue(cfg.Language); ok {
		currentLanguage = lang
	}
}

func clampUserRounds(rounds int) int {
	if rounds < 1 {
		return 1
	}
	if rounds > 99 {
		return 99
	}
	return rounds
}

func clampProjectileReentry(value int) int {
	if value < 0 || value > 2 {
		return 1
	}
	return value
}

func clampPalmCount(value int) int {
	if value < -1 || value > 2 {
		return -1
	}
	return value
}

func clampCloudAggression(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func languageConfigValue(lang languageID) string {
	if lang == languageEnglish {
		return "en"
	}
	return "de"
}

func parseLanguageConfigValue(value string) (languageID, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "de", "german", "deutsch":
		return languageGerman, true
	case "en", "english", "englisch":
		return languageEnglish, true
	default:
		return 0, false
	}
}
