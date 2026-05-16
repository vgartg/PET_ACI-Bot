package executor

import (
	"errors"
	"testing"
)

func TestOpenFileValidation(t *testing.T) {
	t.Parallel()
	s := NewSystem()
	if err := s.OpenFile("   "); !errors.Is(err, ErrEmptyPath) {
		t.Errorf("OpenFile empty: got %v, want %v", err, ErrEmptyPath)
	}
}

func TestOpenURLValidation(t *testing.T) {
	t.Parallel()
	s := NewSystem()
	if err := s.OpenURL(""); !errors.Is(err, ErrEmptyURL) {
		t.Errorf("OpenURL empty: got %v, want %v", err, ErrEmptyURL)
	}
}

func TestKillProcessValidation(t *testing.T) {
	t.Parallel()
	s := NewSystem()
	if err := s.KillProcess(""); !errors.Is(err, ErrEmptyProcessName) {
		t.Errorf("KillProcess empty: got %v, want %v", err, ErrEmptyProcessName)
	}
}

func TestBrowserCommand(t *testing.T) {
	t.Parallel()
	name, args := browserCommand("https://example.com")
	if name == "" {
		t.Fatal("expected non-empty command")
	}
	if len(args) == 0 || args[len(args)-1] != "https://example.com" {
		t.Errorf("unexpected args: %v", args)
	}
}
