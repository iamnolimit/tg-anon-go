package plugins

import (
	"context"
	"log"
	"strings"

	"tg-anon-go/constants"
	"tg-anon-go/matcher"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Manager mengelola semua plugin
type Manager struct {
	plugins     []Plugin
	commands    map[string]Plugin
	adminPlugin *AdminPlugin
	matcher     *matcher.Matcher
}

// NewManager membuat instance Manager baru
func NewManager() *Manager {
	return &Manager{
		plugins:  make([]Plugin, 0),
		commands: make(map[string]Plugin),
	}
}

// Register mendaftarkan plugin
func (m *Manager) Register(plugin Plugin) {
	m.plugins = append(m.plugins, plugin)

	// Map commands ke plugin
	for _, cmd := range plugin.Commands() {
		m.commands[cmd] = plugin
		log.Printf("📦 Registered command: /%s from plugin: %s", cmd, plugin.Name())
	}

	log.Printf("✅ Plugin loaded: %s", plugin.Name())
}

// SetMatcher sets the matcher instance
func (m *Manager) SetMatcher(mch *matcher.Matcher) {
	m.matcher = mch

	// Pass matcher to ChatPlugin
	for _, plugin := range m.plugins {
		if chatPlugin, ok := plugin.(*ChatPlugin); ok {
			chatPlugin.SetMatcher(mch)
			break
		}
	}

	log.Println("✅ Matcher instance set in plugin manager and ChatPlugin")
}

// GetMatcher gets the matcher instance
func (m *Manager) GetMatcher() *matcher.Matcher {
	return m.matcher
}

// HandleUpdate menangani update dari Telegram
func (m *Manager) HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// Handle callback query (inline keyboard)
	if update.CallbackQuery != nil {
		m.handleCallback(bot, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	message := update.Message

	// Ignore messages from log group to prevent loop
	if message.Chat.ID == constants.LogGroupID {
		log.Printf("🚫 Ignoring message from log group: %d", message.Chat.ID)
		return
	}

	// Ignore group messages (only handle private chats)
	if !message.Chat.IsPrivate() {
		log.Printf("🚫 Ignoring group message from chat: %d", message.Chat.ID)
		return
	}

	// Check if user is owner - owners bypass fsub
	isOwner := false
	for _, ownerID := range constants.OwnerIDs {
		if message.From.ID == ownerID {
			isOwner = true
			break
		}
	}

	// Check fsub (force subscribe) - except for /start command and owners
	if !isOwner {
		if message.IsCommand() {
			command := strings.ToLower(message.Command())
			if command != "start" {
				ctx := context.Background()
				allowed, channel := CheckFsub(ctx, bot, message.From.ID)
				if !allowed {
					SendFsubPrompt(bot, message.Chat.ID, channel)
					return
				}
			}
		} else {
			// Check fsub for regular messages too
			ctx := context.Background()
			allowed, channel := CheckFsub(ctx, bot, message.From.ID)
			if !allowed {
				SendFsubPrompt(bot, message.Chat.ID, channel)
				return
			}
		}
	}

	// Handle command
	if message.IsCommand() {
		command := strings.ToLower(message.Command())

		if plugin, exists := m.commands[command]; exists {
			if err := plugin.HandleCommand(bot, message, command); err != nil {
				log.Printf("Error handling command /%s: %v", command, err)
			}
			return
		}

		// Unknown command
		log.Printf("Unknown command: /%s", command)
		return
	}

	// Handle regular message
	for _, plugin := range m.plugins {
		if plugin.CanHandleMessage(message) {
			if err := plugin.HandleMessage(bot, message); err != nil {
				log.Printf("Error handling message in plugin %s: %v", plugin.Name(), err)
			}
			return
		}
	}
}

// handleCallback menangani callback query dari inline keyboard
func (m *Manager) handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	// Handle fsub verification callback first
	if callback.Data == "fsub_verify" {
		m.handleFsubVerify(bot, callback)
		return
	}

	// Delegate to plugins
	for _, plugin := range m.plugins {
		if plugin.CanHandleCallback(callback.Data) {
			if err := plugin.HandleCallbackQuery(bot, callback); err != nil {
				log.Printf("Error handling callback in plugin %s: %v", plugin.Name(), err)
			}
			return
		}
	}
}

// LoadDefaultPlugins memuat plugin default
func (m *Manager) LoadDefaultPlugins() {
	m.Register(NewStartPlugin())
	m.Register(NewChatPlugin())

	// Register admin plugin
	m.adminPlugin = NewAdminPlugin()
	m.Register(m.adminPlugin)

	log.Println("✅ All default plugins loaded")
}

// handleFsubVerify handles fsub verification callback
func (m *Manager) handleFsubVerify(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	ctx := context.Background()
	userID := callback.From.ID

	// Check if user is now a member
	allowed, channel := CheckFsub(ctx, bot, userID)

	// Answer callback
	var callbackText string
	if allowed {
		callbackText = "✅ Verifikasi berhasil!"
	} else {
		callbackText = "❌ Belum join channel"
	}

	callbackResponse := tgbotapi.NewCallback(callback.ID, callbackText)
	bot.Send(callbackResponse)

	// Delete the prompt message
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	bot.Send(deleteMsg)

	// Send result message
	var msg tgbotapi.MessageConfig
	if allowed {
		msg = tgbotapi.NewMessage(callback.Message.Chat.ID, constants.MsgFsubVerified)
	} else {
		msg = tgbotapi.NewMessage(callback.Message.Chat.ID, constants.MsgFsubNotJoined)
		// Resend prompt
		SendFsubPrompt(bot, callback.Message.Chat.ID, channel)
	}
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// GetAdminPlugin mengembalikan admin plugin untuk akses ads
func (m *Manager) GetAdminPlugin() *AdminPlugin {
	return m.adminPlugin
}
