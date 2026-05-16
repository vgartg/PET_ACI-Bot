package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vgartg/aci-bot/internal/bot"
	"github.com/vgartg/aci-bot/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	b, err := bot.New(cfg, logger)
	if err != nil {
		logger.Error("initialize bot", "err", err)
		os.Exit(1)
	}

	logger.Info("bot started")
	if err := b.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("run loop", "err", err)
		os.Exit(1)
	}
	logger.Info("bot stopped")
}
