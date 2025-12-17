package plugins

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Manager mengelola semua plugin
type Manager struct {
	plugins     []Plugin
	commands    map[string]Plugin
	adminPlugin *AdminPlugin
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

// GetAdminPlugin mengembalikan admin plugin untuk akses ads
func (m *Manager) GetAdminPlugin() *AdminPlugin {
	return m.adminPlugin
}
