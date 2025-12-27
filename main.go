package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tg-anon-go/databases"
	"tg-anon-go/matcher"
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
	} // Set debug mode
	if os.Getenv("BOT_DEBUG") == "true" {
		bot.Debug = true
	}
	log.Printf("🤖 Bot authorized on account @%s", bot.Self.UserName)

	// Initialize Redis matcher
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		// Use default Redis URL if not set
		redisURL = "redis://default:5NQHBzWOhwHrczAy8SfFtqCCoPcHVTzn@redis-12448.crce194.ap-seast-1-1.ec2.cloud.redislabs.com:12448"
		log.Println("⚠️ REDIS_URL not set, using default Redis instance")
	}

	redisMatcher, err := matcher.NewMatcher(bot, redisURL)
	if err != nil {
		log.Fatalf("Failed to initialize Redis matcher: %v", err)
	}
	defer redisMatcher.Stop()

	// Start matcher
	redisMatcher.Start()

	// Initialize plugin manager
	pluginManager := plugins.NewManager()
	pluginManager.LoadDefaultPlugins()

	// Set matcher instance to plugin manager
	pluginManager.SetMatcher(redisMatcher)

	// Check run mode
	port := os.Getenv("PORT")
	usePolling := os.Getenv("USE_POLLING") == "true"
	webhookURL := os.Getenv("WEBHOOK_URL")

	// Self-ping is DISABLED by default to save dyno hours
	// Use external services like UptimeRobot or cron-job.org instead
	// Set ENABLE_SELF_PING=true to enable built-in self-ping
	enableSelfPing := os.Getenv("ENABLE_SELF_PING") == "true"
	appURL := os.Getenv("APP_URL")

	// If USE_POLLING is true or no PORT (local), use polling mode
	if usePolling || port == "" {
		// Keep HTTP server alive for Heroku health check if PORT is set
		if port != "" {
			go startHealthServer(port)
			// Start self-ping only if explicitly enabled
			if enableSelfPing {
				startSelfPing(appURL)
			} else {
				log.Println("💡 Self-ping disabled. Use external service (UptimeRobot/cron-job.org) to keep dyno awake")
			}
		}
		runPollingMode(bot, pluginManager)
	} else if webhookURL != "" {
		// Heroku mode with webhook
		runWebhookMode(bot, pluginManager, port, webhookURL, botToken)
	} else {
		// PORT set but no webhook URL - use polling with health server
		log.Println("⚠️ PORT is set but WEBHOOK_URL is empty. Using polling mode...")
		go startHealthServer(port)
		// Start self-ping only if explicitly enabled
		if enableSelfPing {
			startSelfPing(appURL)
		} else {
			log.Println("💡 Self-ping disabled. Use external service (UptimeRobot/cron-job.org) to keep dyno awake")
		}
		runPollingMode(bot, pluginManager)
	}
}

// startHealthServer starts a simple HTTP server for Heroku health checks
func startHealthServer(port string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bot is running!"))
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	log.Printf("🌐 Health check server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("Health server error: %v", err)
	}
}

// startSelfPing starts a self-ping routine to keep the Heroku dyno awake
// NOTE: This uses dyno hours! Consider using external services instead:
// - UptimeRobot (free): https://uptimerobot.com
// - cron-job.org (free): https://cron-job.org
func startSelfPing(appURL string) {
	if appURL == "" {
		log.Println("⚠️ APP_URL not set, self-ping disabled")
		return
	}

	// Ping every 28 minutes (Heroku idles after 30 min)
	ticker := time.NewTicker(28 * time.Minute)
	go func() {
		for range ticker.C {
			resp, err := http.Get(appURL + "/health")
			if err != nil {
				log.Printf("❌ Self-ping failed: %v", err)
			} else {
				resp.Body.Close()
				log.Printf("✅ Self-ping successful")
			}
		}
	}()
	log.Printf("🔄 Self-ping enabled for %s (every 28 min)", appURL)
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
