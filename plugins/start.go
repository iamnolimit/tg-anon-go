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
	return []string{constants.CmdStart, constants.CmdHelp, constants.CmdProfile, constants.CmdEditProfile}
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
	case constants.CmdProfile:
		return p.handleProfile(ctx, bot, chatID, userID)
	case constants.CmdEditProfile:
		return p.handleEditProfile(ctx, bot, chatID, userID)
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
	gender, _ := databases.GetVar(ctx, userID, constants.VarGender)
	location, _ := databases.GetVar(ctx, userID, constants.VarLocation)
	totalChats, _ := databases.GetVarInt(ctx, userID, constants.VarTotalChats)
	totalMessages, _ := databases.GetVarInt(ctx, userID, constants.VarTotalMessages)

	if name == "" {
		name = "Belum diisi"
	}
	if age == "" {
		age = "Belum diisi"
	}
	if gender == "" {
		gender = "Belum diisi"
	}
	if location == "" {
		location = "Belum diisi"
	}

	msg := fmt.Sprintf(constants.MsgProfileInfo, name, age, gender, location, totalChats, totalMessages)
	return p.sendMessage(bot, chatID, msg)
}

// HandleMessage menangani pesan saat registrasi
func (p *StartPlugin) HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	ctx := context.Background()
	chatID := message.Chat.ID
	userID := message.From.ID
	text := message.Text

	// Check edit state first
	editState, _ := databases.GetVar(ctx, userID, constants.VarEditState)
	if editState != "" && editState != constants.EditStateNone {
		switch editState {
		case constants.EditStateName:
			return p.handleEditNameInput(ctx, bot, chatID, userID, text)
		case constants.EditStateAge:
			return p.handleEditAgeInput(ctx, bot, chatID, userID, text)
		case constants.EditStateLocation:
			if message.Location != nil {
				return p.handleEditLocationInput(ctx, bot, chatID, userID, message.Location)
			}
			return p.sendMessage(bot, chatID, constants.MsgRegInvalidLocation)
		}
	}

	// Get registration state
	regState, _ := databases.GetVar(ctx, userID, constants.VarRegState)
		switch regState {
	case constants.RegStateAskName:
		return p.handleNameInput(ctx, bot, chatID, userID, text)
	case constants.RegStateAskAge:
		return p.handleAgeInput(ctx, bot, chatID, userID, text)
	case constants.RegStateAskGender:
		// Gender will be handled by callback
		return p.sendMessage(bot, chatID, constants.MsgRegInvalidGender)
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
	databases.SetVar(ctx, userID, constants.VarRegState, constants.RegStateAskGender)

	msg := fmt.Sprintf(constants.MsgRegAskGender, ageStr)
	return p.sendMessageWithGenderButtons(bot, chatID, msg)
}

// sendMessageWithGenderButtons mengirim pesan dengan tombol pilihan gender
func (p *StartPlugin) sendMessageWithGenderButtons(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	
	// Create inline keyboard with gender options
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👨 "+constants.GenderMale, "gender_male"),
			tgbotapi.NewInlineKeyboardButtonData("👩 "+constants.GenderFemale, "gender_female"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧑 "+constants.GenderOther, "gender_other"),
		),
	)
	msg.ReplyMarkup = keyboard
	
	_, err := bot.Send(msg)
	return err
}

// handleGenderSelection menangani pilihan gender
func (p *StartPlugin) handleGenderSelection(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, gender string) error {
	// Save gender
	databases.SetVar(ctx, userID, constants.VarGender, gender)
	databases.SetVar(ctx, userID, constants.VarRegState, constants.RegStateAskLocation)

	return p.sendMessageWithLocationButton(bot, chatID, constants.MsgRegAskLocation)
}

// handleLocationShare menangani share lokasi
func (p *StartPlugin) handleLocationShare(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, location *tgbotapi.Location) error {
	// Save coordinates
	databases.SetVar(ctx, userID, constants.VarLatitude, location.Latitude)
	databases.SetVar(ctx, userID, constants.VarLongitude, location.Longitude)
	
	// Format location name with coordinates
	locationName := fmt.Sprintf("📍 %.4f, %.4f", location.Latitude, location.Longitude)
	databases.SetVar(ctx, userID, constants.VarLocation, locationName)
	
	// Complete registration
	databases.SetVar(ctx, userID, constants.VarRegState, constants.RegStateDone)
	databases.SetVar(ctx, userID, constants.VarIsRegistered, true)

	// Get saved data for confirmation
	name, _ := databases.GetVar(ctx, userID, constants.VarName)
	age, _ := databases.GetVar(ctx, userID, constants.VarAge)
	gender, _ := databases.GetVar(ctx, userID, constants.VarGender)

	msg := fmt.Sprintf(constants.MsgRegComplete, name, age, gender, locationName)
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
	userID := message.From.ID
	
	// Check edit state
	editState, _ := databases.GetVar(ctx, userID, constants.VarEditState)
	if editState != "" && editState != constants.EditStateNone {
		return true
	}
	
	// Check registration state
	regState, _ := databases.GetVar(ctx, userID, constants.VarRegState)
	
	// Handle message jika user dalam proses registrasi
	return regState == constants.RegStateAskName || 
		   regState == constants.RegStateAskAge || 
		   regState == constants.RegStateAskGender ||
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
	ctx := context.Background()
	chatID := callback.Message.Chat.ID
	userID := callback.From.ID

	// Handle edit profile callbacks
	if callback.Data == "edit_name" || callback.Data == "edit_age" || 
	   callback.Data == "edit_gender" || callback.Data == "edit_location" || 
	   callback.Data == "edit_cancel" || callback.Data == "edit_gender_male" ||
	   callback.Data == "edit_gender_female" || callback.Data == "edit_gender_other" {
		return p.handleEditCallback(ctx, bot, callback)
	}

	// Handle close welcome button
	if (callback.Data == "close_welcome") {
		// Delete the message
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
		bot.Send(deleteMsg)
		
		// Answer callback to remove loading state
		callbackResponse := tgbotapi.NewCallback(callback.ID, "")
		bot.Send(callbackResponse)
		return nil
	}

	// Handle gender selection
	var gender string
	switch callback.Data {
	case "gender_male":
		gender = constants.GenderMale
	case "gender_female":
		gender = constants.GenderFemale
	case "gender_other":
		gender = constants.GenderOther
	default:
		return nil
	}

	// Answer callback
	callbackResponse := tgbotapi.NewCallback(callback.ID, "✅ Gender dipilih!")
	bot.Send(callbackResponse)

	// Delete the gender selection message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	bot.Send(deleteMsg)

	// Continue to location
	return p.handleGenderSelection(ctx, bot, chatID, userID, gender)
}

// CanHandleCallback mengecek apakah plugin bisa handle callback
func (p *StartPlugin) CanHandleCallback(data string) bool {
	return data == "close_welcome" || 
		   data == "gender_male" || 
		   data == "gender_female" || 
		   data == "gender_other" ||
		   data == "edit_name" ||
		   data == "edit_age" ||
		   data == "edit_gender" ||
		   data == "edit_location" ||
		   data == "edit_cancel" ||
		   data == "edit_gender_male" ||
		   data == "edit_gender_female" ||
		   data == "edit_gender_other"
}

// handleEditProfile menampilkan menu edit profil
func (p *StartPlugin) handleEditProfile(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64) error {
	// Check if user is registered
	isRegistered, _ := databases.GetVarBool(ctx, userID, constants.VarIsRegistered)
	if !isRegistered {
		return p.sendMessage(bot, chatID, constants.MsgNotRegistered)
	}

	msg := tgbotapi.NewMessage(chatID, constants.MsgEditProfile)
	msg.ParseMode = "Markdown"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Edit Nama", "edit_name"),
			tgbotapi.NewInlineKeyboardButtonData("📅 Edit Umur", "edit_age"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥 Edit Gender", "edit_gender"),
			tgbotapi.NewInlineKeyboardButtonData("📍 Edit Lokasi", "edit_location"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Batal", "edit_cancel"),
		),
	)
	msg.ReplyMarkup = keyboard
	
	_, err := bot.Send(msg)
	return err
}

// handleEditNameInput menangani input nama baru
func (p *StartPlugin) handleEditNameInput(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, name string) error {
	if name == "" || len(name) < 2 {
		return p.sendMessage(bot, chatID, "⚠️ Nama terlalu pendek. Silakan masukkan nama yang valid.")
	}

	if len(name) > 50 {
		return p.sendMessage(bot, chatID, "⚠️ Nama terlalu panjang. Maksimal 50 karakter.")
	}

	// Update name
	databases.SetVar(ctx, userID, constants.VarName, name)
	databases.SetVar(ctx, userID, constants.VarEditState, constants.EditStateNone)

	return p.sendMessage(bot, chatID, constants.MsgProfileUpdated)
}

// handleEditAgeInput menangani input umur baru
func (p *StartPlugin) handleEditAgeInput(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, ageStr string) error {
	age, err := strconv.Atoi(ageStr)
	if err != nil || age < 13 || age > 100 {
		return p.sendMessage(bot, chatID, constants.MsgRegInvalidAge)
	}

	// Update age
	databases.SetVar(ctx, userID, constants.VarAge, ageStr)
	databases.SetVar(ctx, userID, constants.VarEditState, constants.EditStateNone)

	return p.sendMessage(bot, chatID, constants.MsgProfileUpdated)
}

// handleEditLocationInput menangani input lokasi baru
func (p *StartPlugin) handleEditLocationInput(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, location *tgbotapi.Location) error {
	// Update coordinates
	databases.SetVar(ctx, userID, constants.VarLatitude, location.Latitude)
	databases.SetVar(ctx, userID, constants.VarLongitude, location.Longitude)
	
	// Format location name
	locationName := fmt.Sprintf("📍 %.4f, %.4f", location.Latitude, location.Longitude)
	databases.SetVar(ctx, userID, constants.VarLocation, locationName)
	
	// Clear edit state
	databases.SetVar(ctx, userID, constants.VarEditState, constants.EditStateNone)

	return p.sendMessageRemoveKeyboard(bot, chatID, constants.MsgProfileUpdated)
}

// handleEditCallback menangani callback dari edit profile
func (p *StartPlugin) handleEditCallback(ctx context.Context, bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) error {
	chatID := callback.Message.Chat.ID
	userID := callback.From.ID

	// Answer callback
	callbackResponse := tgbotapi.NewCallback(callback.ID, "")
	bot.Send(callbackResponse)

	// Delete the edit menu message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	bot.Send(deleteMsg)

	switch callback.Data {
	case "edit_name":
		name, _ := databases.GetVar(ctx, userID, constants.VarName)
		databases.SetVar(ctx, userID, constants.VarEditState, constants.EditStateName)
		msg := fmt.Sprintf(constants.MsgEditName, name)
		return p.sendMessage(bot, chatID, msg)

	case "edit_age":
		age, _ := databases.GetVar(ctx, userID, constants.VarAge)
		databases.SetVar(ctx, userID, constants.VarEditState, constants.EditStateAge)
		msg := fmt.Sprintf(constants.MsgEditAge, age)
		return p.sendMessage(bot, chatID, msg)

	case "edit_gender":
		gender, _ := databases.GetVar(ctx, userID, constants.VarGender)
		databases.SetVar(ctx, userID, constants.VarEditState, constants.EditStateGender)
		msg := fmt.Sprintf(constants.MsgEditGender, gender)
		return p.sendEditGenderButtons(bot, chatID, msg)

	case "edit_location":
		location, _ := databases.GetVar(ctx, userID, constants.VarLocation)
		databases.SetVar(ctx, userID, constants.VarEditState, constants.EditStateLocation)
		msg := fmt.Sprintf(constants.MsgEditLocation, location)
		return p.sendMessageWithLocationButton(bot, chatID, msg)

	case "edit_cancel":
		databases.SetVar(ctx, userID, constants.VarEditState, constants.EditStateNone)
		return p.sendMessage(bot, chatID, constants.MsgEditCancelled)

	case "edit_gender_male":
		return p.handleEditGenderSelection(ctx, bot, chatID, userID, constants.GenderMale, callback)
	case "edit_gender_female":
		return p.handleEditGenderSelection(ctx, bot, chatID, userID, constants.GenderFemale, callback)
	case "edit_gender_other":
		return p.handleEditGenderSelection(ctx, bot, chatID, userID, constants.GenderOther, callback)
	}

	return nil
}

// sendEditGenderButtons mengirim tombol pilihan gender untuk edit
func (p *StartPlugin) sendEditGenderButtons(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👨 "+constants.GenderMale, "edit_gender_male"),
			tgbotapi.NewInlineKeyboardButtonData("👩 "+constants.GenderFemale, "edit_gender_female"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧑 "+constants.GenderOther, "edit_gender_other"),
		),
	)
	msg.ReplyMarkup = keyboard
	
	_, err := bot.Send(msg)
	return err
}

// handleEditGenderSelection menangani pilihan gender saat edit
func (p *StartPlugin) handleEditGenderSelection(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, gender string, callback *tgbotapi.CallbackQuery) error {
	// Update gender
	databases.SetVar(ctx, userID, constants.VarGender, gender)
	databases.SetVar(ctx, userID, constants.VarEditState, constants.EditStateNone)

	// Answer callback
	callbackResponse := tgbotapi.NewCallback(callback.ID, "✅ Gender diupdate!")
	bot.Send(callbackResponse)

	// Delete the gender selection message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	bot.Send(deleteMsg)

	return p.sendMessage(bot, chatID, constants.MsgProfileUpdated)
}
