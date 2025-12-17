package plugins

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Plugin interface yang harus diimplementasikan oleh semua plugin
type Plugin interface {
	// Name mengembalikan nama plugin
	Name() string
	
	// Commands mengembalikan daftar command yang ditangani plugin
	Commands() []string
	
	// HandleCommand menangani command
	HandleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, command string) error
	
	// HandleMessage menangani pesan biasa (non-command)
	HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error
	
	// CanHandleMessage mengecek apakah plugin bisa handle message ini
	CanHandleMessage(message *tgbotapi.Message) bool

	// HandleCallbackQuery menangani callback dari inline keyboard
	HandleCallbackQuery(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) error

	// CanHandleCallback mengecek apakah plugin bisa handle callback ini
	CanHandleCallback(data string) bool
}

// BasePlugin struct dasar untuk plugin
type BasePlugin struct{}

// HandleMessage implementasi default - tidak handle apapun
func (b *BasePlugin) HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	return nil
}

// CanHandleMessage implementasi default - tidak handle message
func (b *BasePlugin) CanHandleMessage(message *tgbotapi.Message) bool {
	return false
}

// HandleCallbackQuery implementasi default - tidak handle callback
func (b *BasePlugin) HandleCallbackQuery(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) error {
	return nil
}

// CanHandleCallback implementasi default - tidak handle callback
func (b *BasePlugin) CanHandleCallback(data string) bool {
	return false
}
