package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
		check   func(*testing.T, Config)
	}{
		{
			name: "valid minimal",
			env: map[string]string{
				EnvToken:  "abc",
				EnvKey:    "secret",
				EnvHostID: "12345",
			},
			check: func(t *testing.T, c Config) {
				if c.Token != "abc" {
					t.Errorf("token = %q", c.Token)
				}
				if c.HostID != 12345 {
					t.Errorf("hostID = %d", c.HostID)
				}
				if !c.Sticker.Enabled {
					t.Error("sticker should be enabled by default")
				}
			},
		},
		{
			name:    "missing token",
			env:     map[string]string{EnvKey: "k", EnvHostID: "1"},
			wantErr: "ACI_BOT_TOKEN",
		},
		{
			name:    "missing key",
			env:     map[string]string{EnvToken: "t", EnvHostID: "1"},
			wantErr: "ACI_BOT_KEY",
		},
		{
			name:    "missing host id",
			env:     map[string]string{EnvToken: "t", EnvKey: "k"},
			wantErr: "ACI_HOST_ID",
		},
		{
			name: "invalid host id",
			env: map[string]string{
				EnvToken:  "t",
				EnvKey:    "k",
				EnvHostID: "not-a-number",
			},
			wantErr: "must be int64",
		},
		{
			name: "sticker disabled",
			env: map[string]string{
				EnvToken:      "t",
				EnvKey:        "k",
				EnvHostID:     "1",
				EnvStickerOff: "1",
			},
			check: func(t *testing.T, c Config) {
				if c.Sticker.Enabled {
					t.Error("sticker should be disabled")
				}
			},
		},
		{
			name: "custom paths propagate",
			env: map[string]string{
				EnvToken:    "t",
				EnvKey:      "k",
				EnvHostID:   "1",
				EnvChrome:   "C:/chrome.exe",
				EnvSteam:    "C:/steam.exe",
				EnvDiscord:  "C:/discord.exe",
				EnvFaceit:   "C:/faceit.exe",
				EnvFaceitAC: "C:/faceit-ac.exe",
			},
			check: func(t *testing.T, c Config) {
				if c.Apps.Chrome != "C:/chrome.exe" || c.Apps.Steam != "C:/steam.exe" {
					t.Errorf("paths not loaded: %+v", c.Apps)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			getenv := func(k string) string { return tt.env[k] }
			cfg, err := load(getenv)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestBoolFlag(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"":      false,
		"0":     false,
		"false": false,
		"FALSE": false,
		"no":    false,
		"off":   false,
		"1":     true,
		"true":  true,
		"yes":   true,
		"x":     true,
	}
	for in, want := range cases {
		if got := boolFlag(in); got != want {
			t.Errorf("boolFlag(%q) = %v, want %v", in, got, want)
		}
	}
}
