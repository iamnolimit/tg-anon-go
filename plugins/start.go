package plugins

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"tg-anon-go/constants"
	"tg-anon-go/databases"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StartPlugin menangani command /start dan /help
type StartPlugin struct {
	BasePlugin
}

// NewStartPlugin membuat instance StartPlugin baru
func NewStartPlugin() *StartPlugin {
	return &StartPlugin{}
}

// Name mengembalikan nama plugin
func (p *StartPlugin) Name() string {
	return "start"
}

// Commands mengembalikan daftar command yang ditangani
func (p *StartPlugin) Commands() []string {
	return []string{constants.CmdStart, constants.CmdHelp, "profile"}
}

// HandleCommand menangani command /start dan /help
func (p *StartPlugin) HandleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, command string) error {
	ctx := context.Background()
	chatID := message.Chat.ID
	userID := message.From.ID

	switch command {
	case constants.CmdStart:
		return p.handleStart(ctx, bot, chatID, userID, message)
	case constants.CmdHelp:
		return p.sendMessage(bot, chatID, constants.MsgHelp)
	case "profile":
		return p.handleProfile(ctx, bot, chatID, userID)
	}

	return nil
}

// handleStart menangani command /start
func (p *StartPlugin) handleStart(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, message *tgbotapi.Message) error {
	// Register/update user in database
	username := message.From.UserName
	firstName := message.From.FirstName
	
	err := databases.CreateOrUpdateUser(ctx, userID, username, firstName)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return p.sendMessage(bot, chatID, constants.MsgError)
	}

	// Check if user is already registered
	isRegistered, _ := databases.GetVarBool(ctx, userID, constants.VarIsRegistered)
	if isRegistered {
		// User sudah terdaftar, tampilkan welcome dengan buttons
		return p.sendWelcomeWithButtons(bot, chatID)
	}

	// Mulai proses registrasi
	databases.SetVar(ctx, userID, constants.VarRegState, constants.RegStateAskName)
	return p.sendMessage(bot, chatID, constants.MsgRegWelcome)
}

// sendWelcomeWithButtons mengirim pesan welcome dengan inline keyboard
func (p *StartPlugin) sendWelcomeWithButtons(bot *tgbotapi.BotAPI, chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, constants.MsgWelcome)
	msg.ParseMode = "Markdown"
	
	// Create inline keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("👤 Owner", constants.BotOwnerURL),
			tgbotapi.NewInlineKeyboardButtonURL("📢 Channel", constants.BotChannelURL),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("💬 Support", constants.BotSupportURL),
			tgbotapi.NewInlineKeyboardButtonData("❌ Close", "close_welcome"),
		),
	)
	msg.ReplyMarkup = keyboard
	
	_, err := bot.Send(msg)
	return err
}

// handleProfile menampilkan profil user
func (p *StartPlugin) handleProfile(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64) error {
	name, _ := databases.GetVar(ctx, userID, constants.VarName)
	age, _ := databases.GetVar(ctx, userID, constants.VarAge)
	location, _ := databases.GetVar(ctx, userID, constants.VarLocation)
	totalChats, _ := databases.GetVarInt(ctx, userID, constants.VarTotalChats)
	totalMessages, _ := databases.GetVarInt(ctx, userID, constants.VarTotalMessages)

	if name == "" {
		name = "Belum diisi"
	}
	if age == "" {
		age = "Belum diisi"
	}
	if location == "" {
		location = "Belum diisi"
	}

	msg := fmt.Sprintf(constants.MsgProfileInfo, name, age, location, totalChats, totalMessages)
	return p.sendMessage(bot, chatID, msg)
}

// HandleMessage menangani pesan saat registrasi
func (p *StartPlugin) HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	ctx := context.Background()
	chatID := message.Chat.ID
	userID := message.From.ID
	text := message.Text

	// Get registration state
	regState, _ := databases.GetVar(ctx, userID, constants.VarRegState)
	
	switch regState {
	case constants.RegStateAskName:
		return p.handleNameInput(ctx, bot, chatID, userID, text)
	case constants.RegStateAskAge:
		return p.handleAgeInput(ctx, bot, chatID, userID, text)
	case constants.RegStateAskLocation:
		// Check if user sent location
		if message.Location != nil {
			return p.handleLocationShare(ctx, bot, chatID, userID, message.Location)
		}
		return p.sendMessage(bot, chatID, constants.MsgRegInvalidLocation)
	}

	return nil
}

// handleNameInput menangani input nama
func (p *StartPlugin) handleNameInput(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, name string) error {
	if name == "" || len(name) < 2 {
		return p.sendMessage(bot, chatID, "⚠️ Nama terlalu pendek. Silakan masukkan nama yang valid.")
	}

	if len(name) > 50 {
		return p.sendMessage(bot, chatID, "⚠️ Nama terlalu panjang. Maksimal 50 karakter.")
	}

	// Save name
	databases.SetVar(ctx, userID, constants.VarName, name)
	databases.SetVar(ctx, userID, constants.VarRegState, constants.RegStateAskAge)

	msg := fmt.Sprintf(constants.MsgRegAskAge, name)
	return p.sendMessage(bot, chatID, msg)
}

// handleAgeInput menangani input umur
func (p *StartPlugin) handleAgeInput(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, ageStr string) error {
	age, err := strconv.Atoi(ageStr)
	if err != nil || age < 13 || age > 100 {
		return p.sendMessage(bot, chatID, constants.MsgRegInvalidAge)
	}

	// Save age
	databases.SetVar(ctx, userID, constants.VarAge, ageStr)
	databases.SetVar(ctx, userID, constants.VarRegState, constants.RegStateAskLocation)

	msg := fmt.Sprintf(constants.MsgRegAskLocation, ageStr)
	return p.sendMessageWithLocationButton(bot, chatID, msg)
}

// handleLocationShare menangani share lokasi
func (p *StartPlugin) handleLocationShare(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, location *tgbotapi.Location) error {
	// Save coordinates
	databases.SetVar(ctx, userID, constants.VarLatitude, location.Latitude)
	databases.SetVar(ctx, userID, constants.VarLongitude, location.Longitude)
	
	// Get location name using reverse geocoding (simplified - just store coordinates label)
	locationName := fmt.Sprintf("%.4f, %.4f", location.Latitude, location.Longitude)
	databases.SetVar(ctx, userID, constants.VarLocation, locationName)
	
	// Complete registration
	databases.SetVar(ctx, userID, constants.VarRegState, constants.RegStateDone)
	databases.SetVar(ctx, userID, constants.VarIsRegistered, true)

	// Get saved data for confirmation
	name, _ := databases.GetVar(ctx, userID, constants.VarName)
	age, _ := databases.GetVar(ctx, userID, constants.VarAge)

	msg := fmt.Sprintf(constants.MsgRegComplete, name, age, "📍 Lokasi tersimpan")
	return p.sendMessageRemoveKeyboard(bot, chatID, msg)
}

// sendMessageWithLocationButton mengirim pesan dengan tombol request location
func (p *StartPlugin) sendMessageWithLocationButton(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	
	// Create keyboard with location request button
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonLocation("📍 Bagikan Lokasi"),
		),
	)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true
	msg.ReplyMarkup = keyboard
	
	_, err := bot.Send(msg)
	return err
}

// sendMessageRemoveKeyboard mengirim pesan dan menghapus keyboard
func (p *StartPlugin) sendMessageRemoveKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, err := bot.Send(msg)
	return err
}

// CanHandleMessage mengecek apakah plugin bisa handle message
func (p *StartPlugin) CanHandleMessage(message *tgbotapi.Message) bool {
	if message.IsCommand() {
		return false
	}

	ctx := context.Background()
	regState, _ := databases.GetVar(ctx, message.From.ID, constants.VarRegState)
	
	// Handle message jika user dalam proses registrasi
	return regState == constants.RegStateAskName || 
		   regState == constants.RegStateAskAge || 
		   regState == constants.RegStateAskLocation
}

// sendMessage mengirim pesan dengan Markdown
func (p *StartPlugin) sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := bot.Send(msg)
	return err
}

// HandleCallbackQuery menangani callback dari inline keyboard
func (p *StartPlugin) HandleCallbackQuery(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) error {
	if (callback.Data == "close_welcome") {
		// Delete the message
		deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		bot.Send(deleteMsg)
		
		// Answer callback to remove loading state
		callbackResponse := tgbotapi.NewCallback(callback.ID, "")
		bot.Send(callbackResponse)
	}
	return nil
}

// CanHandleCallback mengecek apakah plugin bisa handle callback
func (p *StartPlugin) CanHandleCallback(data string) bool {
	return data == "close_welcome"
}
