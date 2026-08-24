package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr        string
	LogLevel        string
	DwellSeconds    int
	ShutdownSeconds int
}

func Load() Config {
	c := Config{HTTPAddr: ":8087", LogLevel: "info", DwellSeconds: 300, ShutdownSeconds: 10}
	loadYAML(&c, env("CONFIG_FILE", "configs/config.yaml"))
	c.HTTPAddr = env("HTTP_ADDR", c.HTTPAddr)
	c.LogLevel = env("LOG_LEVEL", c.LogLevel)
	c.DwellSeconds = envInt("DWELL_SECONDS", c.DwellSeconds)
	c.ShutdownSeconds = envInt("SHUTDOWN_SECONDS", c.ShutdownSeconds)
	return c
}

// loadYAML accepts the small scalar configuration schema used by this service.
// Environment variables remain the final override so container deployments need no file mount.
func loadYAML(c *Config, path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := strings.TrimSpace(parts[0]), strings.Trim(strings.TrimSpace(parts[1]), "\"")
		switch key {
		case "http_addr":
			if value != "" {
				c.HTTPAddr = value
			}
		case "log_level":
			if value != "" {
				c.LogLevel = value
			}
		case "dwell_seconds":
			if parsed, err := strconv.Atoi(value); err == nil {
				c.DwellSeconds = parsed
			}
		case "shutdown_seconds":
			if parsed, err := strconv.Atoi(value); err == nil {
				c.ShutdownSeconds = parsed
			}
		}
	}
}
func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, ""))
	if err != nil {
		return fallback
	}
	return value
}
