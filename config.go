package main

import (
	"flag"
	"os"
)

type Config struct {
	Port           string
	SongsDir       string
	DBPath         string
	ServerPassword string
	DailyResetHour int
}

func LoadConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Port, "port", envOrDefault("PORT", "8080"), "server port")
	flag.StringVar(&cfg.SongsDir, "songs-dir", envOrDefault("SONGS_DIR", "./songs"), "songs data directory")
	flag.StringVar(&cfg.DBPath, "db-path", envOrDefault("DB_PATH", "./rankings.db"), "SQLite database path")
	flag.StringVar(&cfg.ServerPassword, "password", envOrDefault("SERVER_PASSWORD", ""), "lobby password (optional)")
	flag.IntVar(&cfg.DailyResetHour, "daily-reset-hour", 0, "daily course reset hour (UTC)")
	flag.Parse()

	return cfg
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
