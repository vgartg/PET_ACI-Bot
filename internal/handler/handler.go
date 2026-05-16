package handler

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/url"
	"strings"

	"github.com/vgartg/aci-bot/internal/config"
	"github.com/vgartg/aci-bot/internal/executor"
	"github.com/vgartg/aci-bot/internal/session"
)

type Message struct {
	ChatID     int64
	Username   string
	Text       string
	HasSticker bool
	HasPhoto   bool
}

type Sender interface {
	SendMessage(chatID int64, text string) error
	SendSticker(chatID int64, fileURL string) error
}

type RandSource interface {
	Intn(n int) int
}

type Handler struct {
	cfg      config.Config
	exec     executor.Executor
	sessions *session.Manager
	sender   Sender
	logger   *slog.Logger
	rng      RandSource
}

type Option func(*Handler)

func WithLogger(logger *slog.Logger) Option {
	return func(h *Handler) {
		if logger != nil {
			h.logger = logger
		}
	}
}

func WithRand(r RandSource) Option {
	return func(h *Handler) {
		if r != nil {
			h.rng = r
		}
	}
}

func New(cfg config.Config, exec executor.Executor, sessions *session.Manager, sender Sender, opts ...Option) *Handler {
	h := &Handler{
		cfg:      cfg,
		exec:     exec,
		sessions: sessions,
		sender:   sender,
		logger:   slog.Default(),
		rng:      globalRand{},
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

type globalRand struct{}

func (globalRand) Intn(n int) int { return rand.Intn(n) }

const (
	loginPrompt = "Enter the key to activate the bot.\n\n" +
		"If you see this message twice, the key is incorrect"
	welcomeText = "Key accepted. Welcome to ACI\n\n" +
		"I will remember you until this session ends.\n" +
		"/help - list of all commands\n" +
		"/stop - end the session"
	fallbackText        = "Sorry, I did not understand. Try /help"
	sessionEndedText    = "Session ended"
	hostOnlyText        = "Only the host can run this command"
	shutdownFailedText  = "Failed to issue shutdown command"
	shuttingDownText    = "Shutting down the host machine"
	pathMissingText     = "Application path is not configured"
	openAppFailedText   = "Could not start the application"
	openURLFailedText   = "Could not open the URL, check that it is valid"
	noProcessesText     = "No active processes from this application"
	closeProcessOKMsg   = "Closed the process"
	urlPromptText       = "Enter the URL"
	videoPromptText     = "Enter the name of the video you want to watch"
	emptyVideoQueryText = "Empty query"
)

func (h *Handler) Handle(msg Message) error {
	if strings.TrimSpace(msg.Username) == "" {
		return nil
	}
	if !h.sessions.HasUser(msg.Username) {
		return h.handleLogin(msg)
	}
	return h.handleAuthorized(msg)
}

func (h *Handler) handleLogin(msg Message) error {
	if msg.Text != h.cfg.Key {
		return h.sender.SendMessage(msg.ChatID, loginPrompt)
	}
	h.sessions.Login(msg.Username)
	return h.sender.SendMessage(msg.ChatID, welcomeText)
}

func (h *Handler) handleAuthorized(msg Message) error {
	text := strings.TrimSpace(msg.Text)
	if isCommand(text) {
		return h.dispatch(msg, text)
	}
	switch h.sessions.State(msg.Username) {
	case session.StateWaitURL:
		h.sessions.Reset(msg.Username)
		return h.openCustomURL(msg, text)
	case session.StateWaitVideoName:
		h.sessions.Reset(msg.Username)
		return h.openYouTubeSearch(msg, text)
	}
	if msg.HasSticker || msg.HasPhoto {
		return h.sendRandomSticker(msg)
	}
	return h.sender.SendMessage(msg.ChatID, fallbackText)
}

func isCommand(text string) bool {
	return strings.HasPrefix(text, "/")
}

func (h *Handler) dispatch(msg Message, cmd string) error {
	switch cmd {
	case "/help":
		return h.sender.SendMessage(msg.ChatID, HelpText)
	case "/stop":
		h.sessions.Logout(msg.Username)
		return h.sender.SendMessage(msg.ChatID, sessionEndedText)
	case "/turn_off_pc":
		return h.shutdown(msg)
	case "/open_explorer":
		return h.openApp(msg, "explorer.exe", "Opened File Explorer")
	case "/open_google":
		return h.openApp(msg, h.cfg.Apps.Chrome, "Opened Google Chrome")
	case "/open_youtube":
		return h.openURL(msg, "https://www.youtube.com", "Opened YouTube")
	case "/open_video_by_name":
		h.sessions.SetState(msg.Username, session.StateWaitVideoName)
		return h.sender.SendMessage(msg.ChatID, videoPromptText)
	case "/open_vk":
		return h.openURL(msg, "https://vk.com/feed", "Opened VKontakte")
	case "/open_ya_mus":
		return h.openURL(msg, "https://music.yandex.ru/home", "Opened Yandex Music")
	case "/open_url":
		h.sessions.SetState(msg.Username, session.StateWaitURL)
		return h.sender.SendMessage(msg.ChatID, urlPromptText)
	case "/open_faceit":
		return h.openFaceit(msg)
	case "/open_steam":
		return h.openApp(msg, h.cfg.Apps.Steam, "Opened Steam")
	case "/open_discord":
		return h.openApp(msg, h.cfg.Apps.Discord, "Opened Discord")
	case "/close_google":
		return h.closeProcesses(msg, []string{"chrome"}, "Closed Google Chrome")
	case "/close_faceit":
		return h.closeProcesses(msg, []string{"faceit", "faceitclient"}, "Closed Faceit")
	case "/close_steam":
		return h.closeProcesses(msg, []string{"steam"}, "Closed Steam")
	case "/close_discord":
		return h.closeProcesses(msg, []string{"discord"}, "Closed Discord")
	case "/start":
		return h.sender.SendMessage(msg.ChatID, welcomeText)
	default:
		return h.sender.SendMessage(msg.ChatID, fallbackText)
	}
}

func (h *Handler) shutdown(msg Message) error {
	if msg.ChatID != h.cfg.HostID {
		return h.sender.SendMessage(msg.ChatID, hostOnlyText)
	}
	if err := h.exec.Shutdown(); err != nil {
		h.logger.Error("shutdown failed", "err", err)
		return h.sender.SendMessage(msg.ChatID, shutdownFailedText)
	}
	return h.sender.SendMessage(msg.ChatID, shuttingDownText)
}

func (h *Handler) openApp(msg Message, path, successText string) error {
	if strings.TrimSpace(path) == "" {
		return h.sender.SendMessage(msg.ChatID, pathMissingText)
	}
	if err := h.exec.OpenFile(path); err != nil {
		h.logger.Error("open app failed", "path", path, "err", err)
		return h.sender.SendMessage(msg.ChatID, openAppFailedText)
	}
	return h.sender.SendMessage(msg.ChatID, successText)
}

func (h *Handler) openFaceit(msg Message) error {
	if strings.TrimSpace(h.cfg.Apps.Faceit) == "" {
		return h.sender.SendMessage(msg.ChatID, pathMissingText)
	}
	if err := h.exec.OpenFile(h.cfg.Apps.Faceit); err != nil {
		h.logger.Error("open faceit failed", "err", err)
		return h.sender.SendMessage(msg.ChatID, openAppFailedText)
	}
	if strings.TrimSpace(h.cfg.Apps.FaceitAC) != "" {
		if err := h.exec.OpenFile(h.cfg.Apps.FaceitAC); err != nil {
			h.logger.Warn("open faceit anti-cheat failed", "err", err)
		}
	}
	return h.sender.SendMessage(msg.ChatID, "Opened Faceit")
}

func (h *Handler) openURL(msg Message, raw, successText string) error {
	if err := h.exec.OpenURL(raw); err != nil {
		h.logger.Error("open url failed", "url", raw, "err", err)
		return h.sender.SendMessage(msg.ChatID, openURLFailedText)
	}
	return h.sender.SendMessage(msg.ChatID, successText)
}

func (h *Handler) openCustomURL(msg Message, raw string) error {
	if !validURL(raw) {
		return h.sender.SendMessage(msg.ChatID, openURLFailedText)
	}
	return h.openURL(msg, raw, "Opened your URL")
}

func (h *Handler) openYouTubeSearch(msg Message, query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return h.sender.SendMessage(msg.ChatID, emptyVideoQueryText)
	}
	target := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
	return h.openURL(msg, target, fmt.Sprintf("Searching YouTube for: %s", query))
}

func (h *Handler) closeProcesses(msg Message, names []string, successText string) error {
	killed := 0
	for _, name := range names {
		if err := h.exec.KillProcess(name); err != nil {
			h.logger.Warn("kill process failed", "name", name, "err", err)
			continue
		}
		killed++
	}
	if killed == 0 {
		return h.sender.SendMessage(msg.ChatID, noProcessesText)
	}
	return h.sender.SendMessage(msg.ChatID, successText)
}

func (h *Handler) sendRandomSticker(msg Message) error {
	s := h.cfg.Sticker
	if !s.Enabled {
		return nil
	}
	span := s.Max - s.Min + 1
	if span <= 0 {
		return nil
	}
	n := h.rng.Intn(span) + s.Min
	target := fmt.Sprintf(s.URLPattern, fmt.Sprintf("%03d", n))
	return h.sender.SendSticker(msg.ChatID, target)
}

func validURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return parsed.Host != ""
}
