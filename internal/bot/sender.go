package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type telegramSender struct {
	api      *tgbotapi.BotAPI
	keyboard tgbotapi.ReplyKeyboardMarkup
}

func newTelegramSender(api *tgbotapi.BotAPI) *telegramSender {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("/help")),
	)
	keyboard.ResizeKeyboard = true
	return &telegramSender{api: api, keyboard: keyboard}
}

func (s *telegramSender) SendMessage(chatID int64, text string) error {
	out := tgbotapi.NewMessage(chatID, text)
	out.ReplyMarkup = s.keyboard
	_, err := s.api.Send(out)
	return err
}

func (s *telegramSender) SendSticker(chatID int64, fileURL string) error {
	sticker := tgbotapi.NewSticker(chatID, tgbotapi.FileURL(fileURL))
	_, err := s.api.Send(sticker)
	return err
}
