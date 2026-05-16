package executor

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

var (
	ErrEmptyPath        = errors.New("empty path")
	ErrEmptyURL         = errors.New("empty url")
	ErrEmptyProcessName = errors.New("empty process name")
)

type System struct{}

func NewSystem() *System {
	return &System{}
}

func (s *System) OpenFile(path string) error {
	if strings.TrimSpace(path) == "" {
		return ErrEmptyPath
	}
	return exec.Command(path).Start()
}

func (s *System) OpenURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return ErrEmptyURL
	}
	name, args := browserCommand(rawURL)
	return exec.Command(name, args...).Start()
}

func (s *System) KillProcess(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyProcessName
	}
	image := name
	if !strings.HasSuffix(strings.ToLower(image), ".exe") {
		image += ".exe"
	}
	return exec.Command("taskkill", "/F", "/IM", image).Run()
}

func (s *System) Shutdown() error {
	return exec.Command("shutdown", "/s", "/f", "/t", "0").Start()
}

func browserCommand(rawURL string) (string, []string) {
	switch runtime.GOOS {
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", rawURL}
	case "darwin":
		return "open", []string{rawURL}
	default:
		return "xdg-open", []string{rawURL}
	}
}
