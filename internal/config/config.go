package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Token   string
	Key     string
	HostID  int64
	Apps    AppPaths
	Sticker StickerConfig
}

type AppPaths struct {
	Chrome   string
	Steam    string
	Discord  string
	Faceit   string
	FaceitAC string
}

type StickerConfig struct {
	Enabled    bool
	URLPattern string
	Min        int
	Max        int
}

const (
	EnvToken          = "ACI_BOT_TOKEN"
	EnvKey            = "ACI_BOT_KEY"
	EnvHostID         = "ACI_HOST_ID"
	EnvChrome         = "ACI_PATH_CHROME"
	EnvSteam          = "ACI_PATH_STEAM"
	EnvDiscord        = "ACI_PATH_DISCORD"
	EnvFaceit         = "ACI_PATH_FACEIT"
	EnvFaceitAC       = "ACI_PATH_FACEIT_AC"
	EnvStickerOff     = "ACI_STICKER_DISABLED"
	EnvStickerURL     = "ACI_STICKER_URL"
	defaultStickerURL = "https://chpic.su/_data/stickers/k/kisiiiiiii/kisiiiiiii_%s.webp?v=1693179002"
)

func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	token := strings.TrimSpace(getenv(EnvToken))
	if token == "" {
		return Config{}, fmt.Errorf("%s is required", EnvToken)
	}
	key := strings.TrimSpace(getenv(EnvKey))
	if key == "" {
		return Config{}, fmt.Errorf("%s is required", EnvKey)
	}
	hostRaw := strings.TrimSpace(getenv(EnvHostID))
	if hostRaw == "" {
		return Config{}, fmt.Errorf("%s is required", EnvHostID)
	}
	hostID, err := strconv.ParseInt(hostRaw, 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("%s must be int64: %w", EnvHostID, err)
	}

	stickerURL := strings.TrimSpace(getenv(EnvStickerURL))
	if stickerURL == "" {
		stickerURL = defaultStickerURL
	}

	cfg := Config{
		Token:  token,
		Key:    key,
		HostID: hostID,
		Apps: AppPaths{
			Chrome:   getenv(EnvChrome),
			Steam:    getenv(EnvSteam),
			Discord:  getenv(EnvDiscord),
			Faceit:   getenv(EnvFaceit),
			FaceitAC: getenv(EnvFaceitAC),
		},
		Sticker: StickerConfig{
			Enabled:    !boolFlag(getenv(EnvStickerOff)),
			URLPattern: stickerURL,
			Min:        1,
			Max:        119,
		},
	}
	return cfg, nil
}

func boolFlag(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

var ErrEmptyPath = errors.New("application path is not configured")
