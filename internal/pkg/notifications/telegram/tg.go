// Пакет telegram реализует взаимодействие с Telegram Bot API.
package telegram

import (
	"fmt"
	"strconv"

	"github.com/Erikqwerty/KronosKeeper/internal/pkg/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramBot представляет собой структуру для работы с Telegram ботом.
type TelegramBot struct {
	Token  string
	ChatID int64
	*tgbotapi.BotAPI
}

// NewTelegramBot создает новый объект TelegramBot на основе конфигурации.
func NewTelegramBot(conf *config.Telegram) (*TelegramBot, error) {
	botAPI, err := tgbotapi.NewBotAPI(conf.Token)
	if err != nil {
		return nil, err
	}
	chatID, err := strconv.Atoi(conf.ChatID)
	if err != nil {
		return nil, err
	}
	return &TelegramBot{
		Token:  conf.Token,
		ChatID: int64(chatID),
		BotAPI: botAPI,
	}, nil
}

// Start запускает бота и обрабатывает входящие обновления.
func (tg *TelegramBot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tg.BotAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Text == "start" {
			// Отправляем сообщение всем пользователям, отправившим "start"
			err := tg.SendMessage("KronosKeeperBot запущен")
			if err != nil {
				return fmt.Errorf("ошибка отправки сообщения, %v", err)
			}
		}
	}
	return nil
}

// SendMessage отправляет сообщение в чат, указанный в tg.ChatID.
func (tg *TelegramBot) SendMessage(message string) error {
	msg := tgbotapi.NewMessage(tg.ChatID, message)
	_, err := tg.BotAPI.Send(msg)
	return err
}
