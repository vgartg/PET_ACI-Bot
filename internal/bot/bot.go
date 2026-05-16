package bot

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/vgartg/aci-bot/internal/config"
	"github.com/vgartg/aci-bot/internal/executor"
	"github.com/vgartg/aci-bot/internal/handler"
	"github.com/vgartg/aci-bot/internal/session"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	handler *handler.Handler
	logger  *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("create telegram client: %w", err)
	}
	if logger == nil {
		logger = slog.Default()
	}
	sender := newTelegramSender(api)
	sessions := session.New()
	exec := executor.NewSystem()
	h := handler.New(cfg, exec, sessions, sender, handler.WithLogger(logger))
	return &Bot{api: api, handler: h, logger: logger}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 30
	updates := b.api.GetUpdatesChan(cfg)
	defer b.api.StopReceivingUpdates()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case upd, ok := <-updates:
			if !ok {
				return nil
			}
			if upd.Message == nil {
				continue
			}
			msg := toDomainMessage(upd.Message)
			b.logger.Info("incoming", "user", msg.Username, "text", msg.Text)
			if err := b.handler.Handle(msg); err != nil {
				b.logger.Error("handle update", "err", err)
			}
		}
	}
}

func toDomainMessage(m *tgbotapi.Message) handler.Message {
	username := ""
	if m.From != nil {
		username = m.From.UserName
	}
	return handler.Message{
		ChatID:     m.Chat.ID,
		Username:   username,
		Text:       m.Text,
		HasSticker: m.Sticker != nil,
		HasPhoto:   len(m.Photo) > 0,
	}
}
