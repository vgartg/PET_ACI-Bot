package handler

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/vgartg/aci-bot/internal/config"
	"github.com/vgartg/aci-bot/internal/session"
)

type sentMessage struct {
	ChatID int64
	Text   string
}

type sentSticker struct {
	ChatID  int64
	FileURL string
}

type fakeSender struct {
	mu       sync.Mutex
	messages []sentMessage
	stickers []sentSticker
	textErr  error
}

func (f *fakeSender) SendMessage(chatID int64, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages = append(f.messages, sentMessage{ChatID: chatID, Text: text})
	return f.textErr
}

func (f *fakeSender) SendSticker(chatID int64, fileURL string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stickers = append(f.stickers, sentSticker{ChatID: chatID, FileURL: fileURL})
	return nil
}

func (f *fakeSender) lastText() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.messages) == 0 {
		return ""
	}
	return f.messages[len(f.messages)-1].Text
}

type call struct {
	op  string
	arg string
}

type fakeExecutor struct {
	mu         sync.Mutex
	calls      []call
	openErr    error
	urlErr     error
	killErr    error
	shutErr    error
	killByName map[string]error
}

func (f *fakeExecutor) OpenFile(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call{op: "open_file", arg: path})
	return f.openErr
}

func (f *fakeExecutor) OpenURL(rawURL string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call{op: "open_url", arg: rawURL})
	return f.urlErr
}

func (f *fakeExecutor) KillProcess(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call{op: "kill", arg: name})
	if err, ok := f.killByName[name]; ok {
		return err
	}
	return f.killErr
}

func (f *fakeExecutor) Shutdown() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call{op: "shutdown"})
	return f.shutErr
}

func (f *fakeExecutor) opCount(op string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if c.op == op {
			n++
		}
	}
	return n
}

type fixedRand struct{ value int }

func (f fixedRand) Intn(int) int { return f.value }

func newTestHandler(t *testing.T, paths config.AppPaths) (*Handler, *fakeSender, *fakeExecutor, *session.Manager) {
	t.Helper()
	cfg := config.Config{
		Token:  "test-token",
		Key:    "secret",
		HostID: 100,
		Apps:   paths,
		Sticker: config.StickerConfig{
			Enabled:    true,
			URLPattern: "https://example.com/sticker_%s.webp",
			Min:        1,
			Max:        119,
		},
	}
	sender := &fakeSender{}
	exec := &fakeExecutor{}
	sessions := session.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(cfg, exec, sessions, sender, WithLogger(logger), WithRand(fixedRand{value: 41}))
	return h, sender, exec, sessions
}

func TestLoginFlow(t *testing.T) {
	t.Parallel()
	h, sender, _, sessions := newTestHandler(t, config.AppPaths{})

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "hi"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sender.lastText(), "Enter the key") {
		t.Errorf("expected login prompt, got %q", sender.lastText())
	}
	if sessions.HasUser("alice") {
		t.Error("user should not be logged in yet")
	}

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "secret"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sessions.HasUser("alice") {
		t.Error("user should be logged in after correct key")
	}
	if !strings.Contains(sender.lastText(), "Welcome to ACI") {
		t.Errorf("expected welcome, got %q", sender.lastText())
	}
}

func TestHelpCommand(t *testing.T) {
	t.Parallel()
	h, sender, _, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "/help"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sender.lastText(), "Available commands") {
		t.Errorf("expected help text, got %q", sender.lastText())
	}
}

func TestStopCommand(t *testing.T) {
	t.Parallel()
	h, _, _, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "/stop"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions.HasUser("alice") {
		t.Error("user should be logged out after /stop")
	}
}

func TestTurnOffPCHostOnly(t *testing.T) {
	t.Parallel()
	h, sender, exec, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 999, Username: "alice", Text: "/turn_off_pc"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.opCount("shutdown") != 0 {
		t.Error("shutdown should not be called from non-host chat")
	}
	if !strings.Contains(sender.lastText(), "Only the host") {
		t.Errorf("expected host-only message, got %q", sender.lastText())
	}

	if err := h.Handle(Message{ChatID: 100, Username: "alice", Text: "/turn_off_pc"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.opCount("shutdown") != 1 {
		t.Errorf("shutdown call count = %d, want 1", exec.opCount("shutdown"))
	}
}

func TestOpenAppMissingPath(t *testing.T) {
	t.Parallel()
	h, sender, exec, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "/open_steam"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.opCount("open_file") != 0 {
		t.Error("open_file should not be called when path is empty")
	}
	if !strings.Contains(sender.lastText(), "path is not configured") {
		t.Errorf("got %q", sender.lastText())
	}
}

func TestOpenAppSuccess(t *testing.T) {
	t.Parallel()
	h, sender, exec, sessions := newTestHandler(t, config.AppPaths{Steam: "C:/steam.exe"})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "/open_steam"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.opCount("open_file") != 1 {
		t.Error("open_file should be called once")
	}
	if !strings.Contains(sender.lastText(), "Opened Steam") {
		t.Errorf("got %q", sender.lastText())
	}
}

func TestOpenFaceitWithAntiCheat(t *testing.T) {
	t.Parallel()
	h, _, exec, sessions := newTestHandler(t, config.AppPaths{
		Faceit:   "C:/faceit.exe",
		FaceitAC: "C:/faceit-ac.exe",
	})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "/open_faceit"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.opCount("open_file") != 2 {
		t.Errorf("expected 2 open_file calls, got %d", exec.opCount("open_file"))
	}
}

func TestOpenURLs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		cmd        string
		wantURL    string
		wantInText string
	}{
		{cmd: "/open_youtube", wantURL: "https://www.youtube.com", wantInText: "YouTube"},
		{cmd: "/open_vk", wantURL: "https://vk.com/feed", wantInText: "VKontakte"},
		{cmd: "/open_ya_mus", wantURL: "https://music.yandex.ru/home", wantInText: "Yandex Music"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.cmd, func(t *testing.T) {
			t.Parallel()
			h, sender, exec, sessions := newTestHandler(t, config.AppPaths{})
			sessions.Login("alice")

			if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: tt.cmd}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exec.calls[0].op != "open_url" || exec.calls[0].arg != tt.wantURL {
				t.Errorf("got %+v, want url %q", exec.calls[0], tt.wantURL)
			}
			if !strings.Contains(sender.lastText(), tt.wantInText) {
				t.Errorf("got %q", sender.lastText())
			}
		})
	}
}

func TestOpenCustomURLFlow(t *testing.T) {
	t.Parallel()
	h, sender, exec, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "/open_url"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions.State("alice") != session.StateWaitURL {
		t.Fatalf("state = %v, want StateWaitURL", sessions.State("alice"))
	}
	if !strings.Contains(sender.lastText(), "Enter the URL") {
		t.Errorf("got %q", sender.lastText())
	}

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "https://example.org"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions.State("alice") != session.StateIdle {
		t.Error("state should reset to idle after consuming URL")
	}
	if exec.opCount("open_url") != 1 {
		t.Errorf("open_url calls = %d", exec.opCount("open_url"))
	}
	if exec.calls[len(exec.calls)-1].arg != "https://example.org" {
		t.Errorf("unexpected url: %v", exec.calls[len(exec.calls)-1])
	}
}

func TestOpenCustomURLRejectsInvalid(t *testing.T) {
	t.Parallel()
	h, sender, exec, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")
	sessions.SetState("alice", session.StateWaitURL)

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "not a url"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.opCount("open_url") != 0 {
		t.Error("invalid url should not be opened")
	}
	if !strings.Contains(sender.lastText(), "valid") {
		t.Errorf("expected validation failure message, got %q", sender.lastText())
	}
}

func TestOpenVideoByName(t *testing.T) {
	t.Parallel()
	h, _, exec, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "/open_video_by_name"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions.State("alice") != session.StateWaitVideoName {
		t.Fatal("state should be waiting for video name")
	}

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "go programming"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.opCount("open_url") != 1 {
		t.Errorf("open_url calls = %d", exec.opCount("open_url"))
	}
	last := exec.calls[len(exec.calls)-1].arg
	if !strings.Contains(last, "search_query=go+programming") && !strings.Contains(last, "search_query=go%20programming") {
		t.Errorf("unexpected url: %s", last)
	}
}

func TestCloseProcessNothingKilled(t *testing.T) {
	t.Parallel()
	h, sender, _, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")

	exec := &fakeExecutor{killErr: errors.New("not found")}
	h.exec = exec

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "/close_google"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sender.lastText(), "No active processes") {
		t.Errorf("got %q", sender.lastText())
	}
}

func TestCloseFaceitKillsBoth(t *testing.T) {
	t.Parallel()
	h, sender, exec, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "/close_faceit"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.opCount("kill") != 2 {
		t.Errorf("kill count = %d, want 2", exec.opCount("kill"))
	}
	if !strings.Contains(sender.lastText(), "Closed Faceit") {
		t.Errorf("got %q", sender.lastText())
	}
}

func TestFallbackText(t *testing.T) {
	t.Parallel()
	h, sender, _, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 1, Username: "alice", Text: "blah"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sender.lastText(), "did not understand") {
		t.Errorf("got %q", sender.lastText())
	}
}

func TestStickerEasterEgg(t *testing.T) {
	t.Parallel()
	h, sender, _, sessions := newTestHandler(t, config.AppPaths{})
	sessions.Login("alice")

	if err := h.Handle(Message{ChatID: 1, Username: "alice", HasSticker: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sender.stickers) != 1 {
		t.Fatalf("sticker count = %d", len(sender.stickers))
	}
	if !strings.Contains(sender.stickers[0].FileURL, "042") {
		t.Errorf("expected zero-padded number 042 in url, got %q", sender.stickers[0].FileURL)
	}
}

func TestEmptyUsernameIgnored(t *testing.T) {
	t.Parallel()
	h, sender, _, _ := newTestHandler(t, config.AppPaths{})
	if err := h.Handle(Message{ChatID: 1, Username: "", Text: "hi"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sender.messages) != 0 {
		t.Error("nothing should be sent for empty username")
	}
}

func TestValidURL(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"https://example.com":   true,
		"http://example.com":    true,
		"ftp://example.com":     false,
		"":                      false,
		"not a url":             false,
		"  https://example.com": true,
	}
	for in, want := range cases {
		if got := validURL(in); got != want {
			t.Errorf("validURL(%q) = %v, want %v", in, got, want)
		}
	}
}
