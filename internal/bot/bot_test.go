package bot

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestToDomainMessageBasic(t *testing.T) {
	t.Parallel()
	m := &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 42},
		From: &tgbotapi.User{UserName: "alice"},
		Text: "/help",
	}
	got := toDomainMessage(m)
	if got.ChatID != 42 || got.Username != "alice" || got.Text != "/help" {
		t.Errorf("unexpected mapping: %+v", got)
	}
	if got.HasSticker || got.HasPhoto {
		t.Error("flags should be false")
	}
}

func TestToDomainMessageSticker(t *testing.T) {
	t.Parallel()
	m := &tgbotapi.Message{
		Chat:    &tgbotapi.Chat{ID: 1},
		From:    &tgbotapi.User{UserName: "bob"},
		Sticker: &tgbotapi.Sticker{},
	}
	got := toDomainMessage(m)
	if !got.HasSticker {
		t.Error("HasSticker should be true")
	}
}

func TestToDomainMessagePhoto(t *testing.T) {
	t.Parallel()
	m := &tgbotapi.Message{
		Chat:  &tgbotapi.Chat{ID: 1},
		From:  &tgbotapi.User{UserName: "bob"},
		Photo: []tgbotapi.PhotoSize{{}},
	}
	got := toDomainMessage(m)
	if !got.HasPhoto {
		t.Error("HasPhoto should be true")
	}
}

func TestToDomainMessageEmptyFrom(t *testing.T) {
	t.Parallel()
	m := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}}
	got := toDomainMessage(m)
	if got.Username != "" {
		t.Errorf("username should be empty, got %q", got.Username)
	}
}
