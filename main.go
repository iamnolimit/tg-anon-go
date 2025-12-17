package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"tg-anon-go/databases"
	"tg-anon-go/plugins"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (optional, Heroku uses env vars)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize database
	if err := databases.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer databases.CloseDatabase()

	// Initialize Telegram bot
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN environment variable is not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Set debug mode
	if os.Getenv("BOT_DEBUG") == "true" {
		bot.Debug = true
	}
	log.Printf("🤖 Bot authorized on account @%s", bot.Self.UserName)

	// Initialize plugin manager
	pluginManager := plugins.NewManager()
	pluginManager.LoadDefaultPlugins()

	// Check if running on Heroku (webhook mode) or local (polling mode)
	port := os.Getenv("PORT")
	webhookURL := os.Getenv("WEBHOOK_URL")
	herokuAppName := os.Getenv("HEROKU_APP_NAME") // Auto-set by Heroku if dyno metadata enabled

	// Auto-generate webhook URL if not set but running on Heroku
	if port != "" && webhookURL == "" && herokuAppName != "" {
		webhookURL = "https://" + herokuAppName + ".herokuapp.com"
		log.Printf("📍 Auto-generated WEBHOOK_URL from HEROKU_APP_NAME: %s", webhookURL)
	}

	if port != "" && webhookURL != "" {
		// Heroku mode: use webhook
		runWebhookMode(bot, pluginManager, port, webhookURL, botToken)
	} else if port != "" {
		// Heroku without webhook URL - still try webhook mode with default URL
		log.Println("⚠️ PORT is set but WEBHOOK_URL is empty. Please set WEBHOOK_URL env var.")
		log.Println("💡 Run: heroku config:set WEBHOOK_URL=https://your-app-name.herokuapp.com")
		runPollingMode(bot, pluginManager)
	} else {
		// Local mode: use polling
		runPollingMode(bot, pluginManager)
	}
}

// runWebhookMode menjalankan bot dengan webhook (untuk Heroku)
func runWebhookMode(bot *tgbotapi.BotAPI, pluginManager *plugins.Manager, port, webhookURL, botToken string) {
	// Set webhook
	webhookFullURL := webhookURL + "/webhook/" + botToken
	webhook, err := tgbotapi.NewWebhook(webhookFullURL)
	if err != nil {
		log.Fatalf("Failed to create webhook: %v", err)
	}

	_, err = bot.Request(webhook)
	if err != nil {
		log.Fatalf("Failed to set webhook: %v", err)
	}

	log.Printf("🌐 Webhook set to: %s", webhookFullURL)

	// Get webhook info
	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Printf("Error getting webhook info: %v", err)
	} else {
		if info.LastErrorDate != 0 {
			log.Printf("⚠️ Telegram webhook error: %s", info.LastErrorMessage)
		}
		log.Printf("📊 Webhook info - Pending updates: %d", info.PendingUpdateCount)
	}

	// Set up updates channel from webhook
	updates := bot.ListenForWebhook("/webhook/" + botToken)

	// Start HTTP server in goroutine
	go func() {
		log.Printf("🚀 Starting webhook server on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🚀 Bot is running in WEBHOOK mode...")

	// Main loop
	for {
		select {
		case update := <-updates:
			go pluginManager.HandleUpdate(bot, update)
		case <-sigChan:
			log.Println("\n👋 Shutting down bot...")
			// Remove webhook on shutdown
			bot.Request(tgbotapi.DeleteWebhookConfig{})
			return
		}
	}
}

// runPollingMode menjalankan bot dengan long polling (untuk local development)
func runPollingMode(bot *tgbotapi.BotAPI, pluginManager *plugins.Manager) {
	// Remove any existing webhook
	bot.Request(tgbotapi.DeleteWebhookConfig{})

	// Set up update config
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	// Get updates channel
	updates := bot.GetUpdatesChan(updateConfig)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🚀 Bot is running in POLLING mode... Press Ctrl+C to stop")

	// Main loop
	for {
		select {
		case update := <-updates:
			go pluginManager.HandleUpdate(bot, update)
		case <-sigChan:
			log.Println("\n👋 Shutting down bot...")
			bot.StopReceivingUpdates()
			return
		}
	}
}
