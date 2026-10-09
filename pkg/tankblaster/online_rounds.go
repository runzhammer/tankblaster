package tankblaster

func normalizedOnlineRounds(rounds int) int {
	if rounds < 1 {
		return 1
	}
	if rounds > 99 {
		return 99
	}
	return rounds
}

func loadOnlineRounds(fallback int) int {
	path, err := userConfigPath()
	if err != nil {
		return normalizedOnlineRounds(fallback)
	}
	return loadOnlineRoundsFromPath(path, fallback)
}

func loadOnlineRoundsFromPath(path string, fallback int) int {
	cfg, err := readUserConfigFile(path)
	if err != nil || cfg.OnlineRounds == nil {
		return normalizedOnlineRounds(fallback)
	}
	return normalizedOnlineRounds(*cfg.OnlineRounds)
}

func saveOnlineRounds(rounds int) {
	path, err := userConfigPath()
	if err != nil {
		return
	}
	_ = saveOnlineRoundsToPath(path, rounds)
}

func saveOnlineRoundsToPath(path string, rounds int) error {
	cfg, err := readUserConfigFile(path)
	if err != nil {
		cfg = userConfigFile{}
	}
	normalized := normalizedOnlineRounds(rounds)
	cfg.OnlineRounds = &normalized
	return writeUserConfigFile(path, cfg)
}
