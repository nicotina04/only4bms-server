package main

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Port           string
	SongsDir       string
	DBPath         string
	ServerPassword string
	DailyResetHour int
	MaxPlayers     int
}

func LoadConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Port, "port", envOrDefault("PORT", "8080"), "server port")
	flag.StringVar(&cfg.SongsDir, "songs-dir", envOrDefault("SONGS_DIR", "./songs"), "songs data directory")
	flag.StringVar(&cfg.DBPath, "db-path", envOrDefault("DB_PATH", "./rankings.db"), "SQLite database path")
	flag.StringVar(&cfg.ServerPassword, "password", envOrDefault("SERVER_PASSWORD", ""), "lobby password (optional)")
	flag.IntVar(&cfg.DailyResetHour, "daily-reset-hour", 0, "daily course reset hour (UTC)")
	flag.IntVar(&cfg.MaxPlayers, "max-players", envOrDefaultInt("MAX_PLAYERS", 2), "maximum players per lobby (min 2)")
	flag.Parse()

	// Parse DAILY_RESET_HOUR from env if flag was not explicitly set
	if v := os.Getenv("DAILY_RESET_HOUR"); v != "" {
		if hour, err := strconv.Atoi(v); err == nil && hour >= 0 && hour < 24 {
			cfg.DailyResetHour = hour
		}
	}

	if cfg.MaxPlayers < 2 {
		cfg.MaxPlayers = 2
	}

	return cfg
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
