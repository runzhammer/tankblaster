package server

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Address             string        `yaml:"address"`
		PublicURL           string        `yaml:"public_url"`
		AllowedOrigins      []string      `yaml:"allowed_origins"`
		ReadHeaderTimeout   time.Duration `yaml:"read_header_timeout"`
		ReadTimeout         time.Duration `yaml:"read_timeout"`
		WriteTimeout        time.Duration `yaml:"write_timeout"`
		IdleTimeout         time.Duration `yaml:"idle_timeout"`
		WebSocketReadLimit  int64         `yaml:"websocket_read_limit"`
		MaxConnections      int           `yaml:"max_connections"`
		MaxSessions         int           `yaml:"max_sessions"`
		MaxQueueLength      int           `yaml:"max_queue_length"`
		ErrorChannelTimeout time.Duration `yaml:"error_channel_timeout"`
	} `yaml:"server"`
	Database struct {
		Driver string `yaml:"driver"`
		Path   string `yaml:"path"`
	} `yaml:"database"`
	Sessions struct {
		MaxPlayers         int           `yaml:"max_players"`
		InviteTTL          time.Duration `yaml:"invite_ttl"`
		InviteTokenLength  int           `yaml:"invite_token_length"`
		JoinCodeLength     int           `yaml:"join_code_length"`
		MaxSessionLifetime time.Duration `yaml:"max_session_lifetime"`
		CleanupInterval    time.Duration `yaml:"cleanup_interval"`
	} `yaml:"sessions"`
	Matchmaking struct {
		Enabled                   bool          `yaml:"enabled"`
		InitialRatingRange        int           `yaml:"initial_rating_range"`
		RatingRangeStep           int           `yaml:"rating_range_step"`
		RatingRangeExpandInterval time.Duration `yaml:"rating_range_expand_interval"`
		MaximumRatingRange        int           `yaml:"maximum_rating_range"`
		QueueTimeout              time.Duration `yaml:"queue_timeout"`
	} `yaml:"matchmaking"`
	Rating struct {
		Enabled       bool `yaml:"enabled"`
		InitialRating int  `yaml:"initial_rating"`
		KFactor       int  `yaml:"k_factor"`
		MinimumRating int  `yaml:"minimum_rating"`
	} `yaml:"rating"`
	Score struct {
		WinPoints  int `yaml:"win_points"`
		LossPoints int `yaml:"loss_points"`
		DrawPoints int `yaml:"draw_points"`
	} `yaml:"score"`
	Leaderboard struct {
		DefaultLimit int `yaml:"default_limit"`
		MaximumLimit int `yaml:"maximum_limit"`
	} `yaml:"leaderboard"`
}

func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	cfg.applyDefaults()
	return cfg, nil
}

func DefaultConfig() Config {
	var cfg Config
	cfg.Server.Address = "127.0.0.1:8765"
	cfg.Server.PublicURL = "http://127.0.0.1:8765"
	cfg.Server.ReadHeaderTimeout = 5 * time.Second
	cfg.Server.IdleTimeout = 60 * time.Second
	cfg.Server.WebSocketReadLimit = 64 * 1024
	cfg.Server.MaxConnections = 256
	cfg.Server.MaxSessions = 128
	cfg.Server.MaxQueueLength = 256
	cfg.Server.ErrorChannelTimeout = 250 * time.Millisecond
	cfg.Database.Driver = "sqlite"
	cfg.Database.Path = "tankblaster.db"
	cfg.Sessions.MaxPlayers = 10
	cfg.Sessions.InviteTTL = 24 * time.Hour
	cfg.Sessions.InviteTokenLength = 32
	cfg.Sessions.JoinCodeLength = 8
	cfg.Sessions.MaxSessionLifetime = 2 * time.Hour
	cfg.Sessions.CleanupInterval = time.Minute
	cfg.Matchmaking.Enabled = true
	cfg.Matchmaking.InitialRatingRange = 100
	cfg.Matchmaking.RatingRangeStep = 100
	cfg.Matchmaking.RatingRangeExpandInterval = 15 * time.Second
	cfg.Matchmaking.MaximumRatingRange = 1000
	cfg.Matchmaking.QueueTimeout = 10 * time.Minute
	cfg.Rating.Enabled = true
	cfg.Rating.InitialRating = 1000
	cfg.Rating.KFactor = 32
	cfg.Rating.MinimumRating = 0
	cfg.Score.WinPoints = 100
	cfg.Score.LossPoints = 0
	cfg.Score.DrawPoints = 25
	cfg.Leaderboard.DefaultLimit = 50
	cfg.Leaderboard.MaximumLimit = 100
	return cfg
}

func (cfg *Config) applyDefaults() {
	def := DefaultConfig()
	if cfg.Server.Address == "" {
		cfg.Server.Address = def.Server.Address
	}
	if cfg.Server.PublicURL == "" {
		cfg.Server.PublicURL = def.Server.PublicURL
	}
	if cfg.Server.ReadHeaderTimeout <= 0 {
		cfg.Server.ReadHeaderTimeout = def.Server.ReadHeaderTimeout
	}
	if cfg.Server.ReadTimeout <= 0 {
		cfg.Server.ReadTimeout = def.Server.ReadTimeout
	}
	if cfg.Server.WriteTimeout <= 0 {
		cfg.Server.WriteTimeout = def.Server.WriteTimeout
	}
	if cfg.Server.IdleTimeout <= 0 {
		cfg.Server.IdleTimeout = def.Server.IdleTimeout
	}
	if cfg.Server.WebSocketReadLimit <= 0 {
		cfg.Server.WebSocketReadLimit = def.Server.WebSocketReadLimit
	}
	if cfg.Server.MaxConnections <= 0 {
		cfg.Server.MaxConnections = def.Server.MaxConnections
	}
	if cfg.Server.MaxSessions <= 0 {
		cfg.Server.MaxSessions = def.Server.MaxSessions
	}
	if cfg.Server.MaxQueueLength <= 0 {
		cfg.Server.MaxQueueLength = def.Server.MaxQueueLength
	}
	if cfg.Server.ErrorChannelTimeout <= 0 {
		cfg.Server.ErrorChannelTimeout = def.Server.ErrorChannelTimeout
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = def.Database.Driver
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = def.Database.Path
	}
	if cfg.Sessions.MaxPlayers <= 0 {
		cfg.Sessions.MaxPlayers = def.Sessions.MaxPlayers
	}
	if cfg.Sessions.InviteTTL <= 0 {
		cfg.Sessions.InviteTTL = def.Sessions.InviteTTL
	}
	if cfg.Sessions.InviteTokenLength <= 0 {
		cfg.Sessions.InviteTokenLength = def.Sessions.InviteTokenLength
	}
	if cfg.Sessions.JoinCodeLength <= 0 {
		cfg.Sessions.JoinCodeLength = def.Sessions.JoinCodeLength
	}
	if cfg.Sessions.MaxSessionLifetime <= 0 {
		cfg.Sessions.MaxSessionLifetime = def.Sessions.MaxSessionLifetime
	}
	if cfg.Sessions.CleanupInterval <= 0 {
		cfg.Sessions.CleanupInterval = def.Sessions.CleanupInterval
	}
	if cfg.Matchmaking.InitialRatingRange <= 0 {
		cfg.Matchmaking.InitialRatingRange = def.Matchmaking.InitialRatingRange
	}
	if cfg.Matchmaking.RatingRangeStep <= 0 {
		cfg.Matchmaking.RatingRangeStep = def.Matchmaking.RatingRangeStep
	}
	if cfg.Matchmaking.RatingRangeExpandInterval <= 0 {
		cfg.Matchmaking.RatingRangeExpandInterval = def.Matchmaking.RatingRangeExpandInterval
	}
	if cfg.Matchmaking.MaximumRatingRange <= 0 {
		cfg.Matchmaking.MaximumRatingRange = def.Matchmaking.MaximumRatingRange
	}
	if cfg.Matchmaking.QueueTimeout <= 0 {
		cfg.Matchmaking.QueueTimeout = def.Matchmaking.QueueTimeout
	}
	if cfg.Rating.InitialRating <= 0 {
		cfg.Rating.InitialRating = def.Rating.InitialRating
	}
	if cfg.Rating.KFactor <= 0 {
		cfg.Rating.KFactor = def.Rating.KFactor
	}
	if cfg.Leaderboard.DefaultLimit <= 0 {
		cfg.Leaderboard.DefaultLimit = def.Leaderboard.DefaultLimit
	}
	if cfg.Leaderboard.MaximumLimit <= 0 {
		cfg.Leaderboard.MaximumLimit = def.Leaderboard.MaximumLimit
	}
}
