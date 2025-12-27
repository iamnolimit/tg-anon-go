package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"tg-anon-go/constants"
	"tg-anon-go/databases"
	"tg-anon-go/matcher"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ChatPlugin menangani command search, next, stop dan pesan chat
type ChatPlugin struct {
	BasePlugin
	matcher *matcher.Matcher
}

// NewChatPlugin membuat instance ChatPlugin baru
func NewChatPlugin() *ChatPlugin {
	return &ChatPlugin{}
}

// SetMatcher sets the matcher instance
func (p *ChatPlugin) SetMatcher(m *matcher.Matcher) {
	p.matcher = m
	log.Println("✅ Matcher instance set in ChatPlugin")
}

// Name mengembalikan nama plugin
func (p *ChatPlugin) Name() string {
	return "chat"
}

// Commands mengembalikan daftar command yang ditangani
func (p *ChatPlugin) Commands() []string {
	return []string{constants.CmdSearch, constants.CmdNext, constants.CmdStop, constants.CmdShare}
}

// HandleCommand menangani command chat
func (p *ChatPlugin) HandleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, command string) error {
	ctx := context.Background()
	chatID := message.Chat.ID
	userID := message.From.ID

	switch command {
	case constants.CmdSearch:
		return p.handleSearch(ctx, bot, chatID, userID)
	case constants.CmdNext:
		return p.handleNext(ctx, bot, chatID, userID)
	case constants.CmdStop:
		return p.handleStop(ctx, bot, chatID, userID)
	case constants.CmdShare:
		return p.handleShare(ctx, bot, chatID, userID, message)
	}

	return nil
}

// handleSearch menangani pencarian partner - langsung cari otomatis
func (p *ChatPlugin) handleSearch(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64) error {
	// Check if user is registered
	isRegistered, _ := databases.GetVarBool(ctx, userID, constants.VarIsRegistered)
	if !isRegistered {
		return p.sendMessage(bot, chatID, constants.MsgNotRegistered)
	}

	// Check if user is banned
	isBanned, _ := databases.IsUserBanned(ctx, userID)
	if isBanned {
		return p.sendMessage(bot, chatID, "❌ Kamu telah dibanned dari bot ini.")
	}

	// Get current status using var system
	status, err := databases.GetUserStatus(ctx, userID)
	if err != nil {
		log.Printf("Error getting user status: %v", err)
		return p.sendMessage(bot, chatID, constants.MsgError)
	}

	// Check current status
	switch status {
	case constants.StatusSearching:
		return p.sendMessage(bot, chatID, constants.MsgAlreadySearching)
	case constants.StatusChatting:
		return p.sendMessage(bot, chatID, constants.MsgAlreadyChatting)
	}
	// Update last active
	databases.UpdateLastActive(ctx, userID)

	// Langsung mulai pencarian otomatis (terdekat dulu, lalu random)
	return p.doSearchAuto(ctx, bot, chatID, userID)
}

// doSearchAuto melakukan pencarian otomatis (terdekat dulu, lalu random)
func (p *ChatPlugin) doSearchAuto(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64) error {
	log.Printf("🔍 User %d mencari partner (auto)...", userID)
	
	// Set user status to searching
	err := databases.SetUserStatus(ctx, userID, constants.StatusSearching)
	if err != nil {
		log.Printf("Error setting user status: %v", err)
		return p.sendMessage(bot, chatID, constants.MsgError)
	}

	// Check if user has location - use nearby mode, else random
	var searchMode string
	var lat, lon float64
	
	if databases.HasLocation(ctx, userID) {
		searchMode = constants.SearchModeNearby
		lat, _ = databases.GetVarFloat64(ctx, userID, constants.VarLatitude)
		lon, _ = databases.GetVarFloat64(ctx, userID, constants.VarLongitude)
		log.Printf("📍 User %d has location, using nearby mode", userID)
	} else {
		searchMode = constants.SearchModeRandom
		log.Printf("🎲 User %d no location, using random mode", userID)
	}

	// Store search mode
	databases.SetVar(ctx, userID, constants.VarSearchMode, searchMode)

	// Publish to Redis matcher
	if p.matcher != nil {
		if err := p.matcher.PublishSearch(ctx, userID, searchMode, lat, lon); err != nil {
			log.Printf("Error publishing search to Redis: %v", err)
			return p.sendMessage(bot, chatID, constants.MsgError)
		}
	}

	log.Printf("✅ User %d sekarang status: searching (%s mode, published to Redis)", userID, searchMode)
	return p.sendMessage(bot, chatID, constants.MsgSearching)
}

// HandleCallbackQuery menangani callback dari inline keyboard
func (p *ChatPlugin) HandleCallbackQuery(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) error {
	ctx := context.Background()

	// Handle warn callback (from log group)
	if strings.HasPrefix(callback.Data, constants.CallbackWarnUser) {
		return p.handleWarnCallback(ctx, bot, callback)
	}

	return nil
}

// CanHandleCallback mengecek apakah plugin bisa handle callback ini
func (p *ChatPlugin) CanHandleCallback(data string) bool {
	return strings.HasPrefix(data, constants.CallbackWarnUser)
}

// handleNext skip partner dan cari baru
func (p *ChatPlugin) handleNext(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64) error {
	status, err := databases.GetUserStatus(ctx, userID)
	if err != nil {
		log.Printf("Error getting user status: %v", err)
		return p.sendMessage(bot, chatID, constants.MsgError)
	}

	switch status {
	case constants.StatusIdle:
		return p.sendMessage(bot, chatID, constants.MsgNotChatting)
	case constants.StatusSearching:
		// Cancel search
		err = databases.SetUserStatus(ctx, userID, constants.StatusIdle)
		if err != nil {
			return p.sendMessage(bot, chatID, constants.MsgError)
		}
		// Remove from Redis searching set
		if p.matcher != nil {
			p.matcher.RemoveSearchingUser(ctx, userID)
		}
		return p.sendMessage(bot, chatID, constants.MsgSearchCancelled)
	case constants.StatusChatting:
		// End current chat and search for new partner
		partnerID, _ := databases.GetUserPartner(ctx, userID)
		if partnerID > 0 {
			// Notify partner
			p.sendMessage(bot, partnerID, constants.MsgPartnerLeft)
			
			// Disconnect users
			databases.DisconnectUsers(ctx, userID, partnerID)
		}

		// Search for new partner
		return p.handleSearch(ctx, bot, chatID, userID)
	}

	return nil
}

// handleStop mengakhiri chat
func (p *ChatPlugin) handleStop(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64) error {
	status, err := databases.GetUserStatus(ctx, userID)
	if err != nil {
		log.Printf("Error getting user status: %v", err)
		return p.sendMessage(bot, chatID, constants.MsgError)
	}

	switch status {
	case constants.StatusIdle:
		return p.sendMessage(bot, chatID, constants.MsgNotChatting)
	case constants.StatusSearching:
		// Cancel search
		err = databases.SetUserStatus(ctx, userID, constants.StatusIdle)
		if err != nil {
			return p.sendMessage(bot, chatID, constants.MsgError)
		}
		// Remove from Redis searching set
		if p.matcher != nil {
			p.matcher.RemoveSearchingUser(ctx, userID)
		}
		return p.sendMessage(bot, chatID, constants.MsgSearchCancelled)
	case constants.StatusChatting:
		partnerID, _ := databases.GetUserPartner(ctx, userID)
		if partnerID > 0 {
			// Notify partner
			p.sendMessage(bot, partnerID, constants.MsgPartnerLeft)
			
			// Disconnect users
			databases.DisconnectUsers(ctx, userID, partnerID)
		} else {
			// Reset user status only
			databases.SetUserStatus(ctx, userID, constants.StatusIdle)
		}

		return p.sendMessage(bot, chatID, constants.MsgChatEnded)
	}

	return nil
}

// HandleMessage menangani pesan chat (forward ke partner)
func (p *ChatPlugin) HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	ctx := context.Background()
	userID := message.From.ID

	status, _ := databases.GetUserStatus(ctx, userID)
	if status != constants.StatusChatting {
		return nil
	}

	partnerID, err := databases.GetUserPartner(ctx, userID)
	if err != nil || partnerID == 0 {
		return nil
	}

	// Update last active & increment message count
	databases.UpdateLastActive(ctx, userID)
	databases.IncrementUserTotalMessages(ctx, userID)
	
	// Increment global message count for ads
	currentCount, _ := databases.GetVarInt(ctx, userID, "msg_count_ads")
	currentCount++
	databases.SetVar(ctx, userID, "msg_count_ads", currentCount)
	
	// Check if should send ads (every N messages)
	if currentCount >= constants.AdsIntervalMessages {
		databases.SetVar(ctx, userID, "msg_count_ads", 0)
		p.sendRandomAds(ctx, bot, userID)
	}

	// Forward message based on type
	switch {
	case message.Text != "":
		return p.forwardTextMessage(bot, partnerID, message.Text)
	case message.Photo != nil:
		return p.forwardPhoto(bot, partnerID, message)
	case message.Sticker != nil:
		return p.forwardSticker(bot, partnerID, message)
	case message.Voice != nil:
		return p.forwardVoice(bot, partnerID, message)
	case message.Video != nil:
		return p.forwardVideo(bot, partnerID, message)
	case message.Document != nil:
		return p.forwardDocument(bot, partnerID, message)
	case message.Animation != nil:
		return p.forwardAnimation(bot, partnerID, message)
	}

	return nil
}

// CanHandleMessage mengecek apakah bisa handle message
func (p *ChatPlugin) CanHandleMessage(message *tgbotapi.Message) bool {
	// Handle semua pesan non-command dari user yang sedang chatting
	if message.IsCommand() {
		return false
	}
	
	ctx := context.Background()
	status, err := databases.GetUserStatus(ctx, message.From.ID)
	if err != nil {
		return false
	}
	
	return status == constants.StatusChatting
}

// Forward functions
func (p *ChatPlugin) forwardTextMessage(bot *tgbotapi.BotAPI, partnerID int64, text string) error {
	msg := tgbotapi.NewMessage(partnerID, "💬 "+text)
	_, err := bot.Send(msg)
	return err
}

func (p *ChatPlugin) forwardPhoto(bot *tgbotapi.BotAPI, partnerID int64, message *tgbotapi.Message) error {
	photo := message.Photo[len(message.Photo)-1] // Get highest resolution
	msg := tgbotapi.NewPhoto(partnerID, tgbotapi.FileID(photo.FileID))
	if message.Caption != "" {
		msg.Caption = "💬 " + message.Caption
	}
	sentMsg, err := bot.Send(msg)
	if err != nil {
		return err
	}

	// Log to log group with warn button
	p.logMediaToGroup(bot, message.From.ID, partnerID, "Photo", photo.FileID, sentMsg.MessageID)
	
	return nil
}

func (p *ChatPlugin) forwardSticker(bot *tgbotapi.BotAPI, partnerID int64, message *tgbotapi.Message) error {
	msg := tgbotapi.NewSticker(partnerID, tgbotapi.FileID(message.Sticker.FileID))
	_, err := bot.Send(msg)
	return err
}

func (p *ChatPlugin) forwardVoice(bot *tgbotapi.BotAPI, partnerID int64, message *tgbotapi.Message) error {
	msg := tgbotapi.NewVoice(partnerID, tgbotapi.FileID(message.Voice.FileID))
	_, err := bot.Send(msg)
	return err
}

func (p *ChatPlugin) forwardVideo(bot *tgbotapi.BotAPI, partnerID int64, message *tgbotapi.Message) error {
	msg := tgbotapi.NewVideo(partnerID, tgbotapi.FileID(message.Video.FileID))
	if message.Caption != "" {
		msg.Caption = "💬 " + message.Caption
	}
	sentMsg, err := bot.Send(msg)
	if err != nil {
		return err
	}

	// Log to log group with warn button
	p.logMediaToGroup(bot, message.From.ID, partnerID, "Video", message.Video.FileID, sentMsg.MessageID)
	
	return nil
}

func (p *ChatPlugin) forwardDocument(bot *tgbotapi.BotAPI, partnerID int64, message *tgbotapi.Message) error {
	msg := tgbotapi.NewDocument(partnerID, tgbotapi.FileID(message.Document.FileID))
	if message.Caption != "" {
		msg.Caption = "💬 " + message.Caption
	}
	_, err := bot.Send(msg)
	return err
}

func (p *ChatPlugin) forwardAnimation(bot *tgbotapi.BotAPI, partnerID int64, message *tgbotapi.Message) error {
	msg := tgbotapi.NewAnimation(partnerID, tgbotapi.FileID(message.Animation.FileID))
	if message.Caption != "" {
		msg.Caption = "💬 " + message.Caption
	}
	_, err := bot.Send(msg)
	return err
}

func (p *ChatPlugin) sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := bot.Send(msg)
	return err
}

// handleShare mengirim data akun Telegram user ke partner
func (p *ChatPlugin) handleShare(ctx context.Context, bot *tgbotapi.BotAPI, chatID, userID int64, message *tgbotapi.Message) error {
	// Check if user is chatting
	status, _ := databases.GetUserStatus(ctx, userID)
	if status != constants.StatusChatting {
		return p.sendMessage(bot, chatID, constants.MsgShareNotChatting)
	}

	// Get partner ID
	partnerID, err := databases.GetUserPartner(ctx, userID)
	if err != nil || partnerID == 0 {
		return p.sendMessage(bot, chatID, constants.MsgShareNotChatting)
	}

	// Get user info from message
	firstName := message.From.FirstName
	lastName := message.From.LastName
	username := message.From.UserName

	// Build full name
	fullName := firstName
	if lastName != "" {
		fullName = firstName + " " + lastName
	}

	// Send to partner
	var shareMsg string
	if username != "" {
		shareMsg = fmt.Sprintf(constants.MsgShareReceived, fullName, username, userID)
	} else {
		shareMsg = fmt.Sprintf(constants.MsgShareNoUsername, fullName, userID)
	}

	// Send share info to partner
	p.sendMessage(bot, partnerID, shareMsg)

	// Confirm to user
	return p.sendMessage(bot, chatID, constants.MsgShareSent)
}

// sendRandomAds mengirim ads random ke user
func (p *ChatPlugin) sendRandomAds(ctx context.Context, bot *tgbotapi.BotAPI, userID int64) {
	// Get ads from global var
	adsEnabled, _ := databases.GetGlobalVarBool(ctx, constants.VarGlobalAdsEnabled)
	if !adsEnabled {
		return
	}

	adsJSON, _ := databases.GetGlobalVar(ctx, constants.VarGlobalAds)
	if adsJSON == "" {
		return
	}

	// Parse ads
	var ads []struct {
		ID      int    `json:"id"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(adsJSON), &ads); err != nil || len(ads) == 0 {
		return
	}

	// Random select
	ad := ads[rand.Intn(len(ads))]
	adsMsg := constants.MsgAdsPrefix + ad.Message

	// Send to user
	msg := tgbotapi.NewMessage(userID, adsMsg)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// logMediaToGroup mengirim media ke log group dengan tombol warn
func (p *ChatPlugin) logMediaToGroup(bot *tgbotapi.BotAPI, senderID, partnerID int64, mediaType, fileID string, sentMessageID int) {
	if constants.LogGroupID == 0 {
		return
	}

	// Format log message
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	logText := fmt.Sprintf(constants.MsgLogMedia, senderID, senderID, partnerID, partnerID, mediaType, currentTime)

	// Create warn button callback data: warn_user_{senderID}_{partnerID}_{sentMessageID}
	callbackData := fmt.Sprintf("%s%d_%d_%d", constants.CallbackWarnUser, senderID, partnerID, sentMessageID)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚠️ Warn User", callbackData),
		),
	)

	// Send media with caption to log group
	var msg tgbotapi.Chattable
	switch mediaType {
	case "Photo":
		photoMsg := tgbotapi.NewPhoto(constants.LogGroupID, tgbotapi.FileID(fileID))
		photoMsg.Caption = logText
		photoMsg.ParseMode = "Markdown"
		photoMsg.ReplyMarkup = keyboard
		msg = photoMsg
	case "Video":
		videoMsg := tgbotapi.NewVideo(constants.LogGroupID, tgbotapi.FileID(fileID))
		videoMsg.Caption = logText
		videoMsg.ParseMode = "Markdown"
		videoMsg.ReplyMarkup = keyboard
		msg = videoMsg
	default:
		return
	}

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error logging media to group: %v", err)
	}
}

// handleWarnCallback menangani callback warn dari log group
func (p *ChatPlugin) handleWarnCallback(ctx context.Context, bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) error {
	// Parse callback data: warn_user_{senderID}_{partnerID}_{sentMessageID}
	data := strings.TrimPrefix(callback.Data, constants.CallbackWarnUser)
	parts := strings.Split(data, "_")
	if len(parts) != 3 {
		return fmt.Errorf("invalid warn callback data")
	}

	senderID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return err
	}
	
	partnerID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}
	
	sentMessageID, err := strconv.Atoi(parts[2])
	if err != nil {
		return err
	}

	// Increment warn count
	currentWarns, _ := databases.GetVarInt(ctx, senderID, constants.VarWarnCount)
	newWarns := currentWarns + 1
	databases.SetVar(ctx, senderID, constants.VarWarnCount, newWarns)

	// Delete the media message from partner's chat
	deleteMsg := tgbotapi.NewDeleteMessage(partnerID, sentMessageID)
	bot.Send(deleteMsg)

	// Check if should auto-ban
	if newWarns >= constants.MaxWarnings {
		// Ban user
		databases.SetVar(ctx, senderID, constants.VarIsBanned, true)
		
		// Disconnect if chatting
		status, _ := databases.GetUserStatus(ctx, senderID)
		if status == constants.StatusChatting {
			partnerIDCurrent, _ := databases.GetUserPartner(ctx, senderID)
			if partnerIDCurrent > 0 {
				p.sendMessage(bot, partnerIDCurrent, constants.MsgPartnerLeft)
				databases.DisconnectUsers(ctx, senderID, partnerIDCurrent)
			}
		}

		// Notify user about ban
		bannedMsg := fmt.Sprintf(constants.MsgWarnedBanned, newWarns)
		p.sendMessage(bot, senderID, bannedMsg)

		// Update callback answer and log message
		callbackResponse := tgbotapi.NewCallback(callback.ID, fmt.Sprintf("User %d telah dibanned!", senderID))
		bot.Send(callbackResponse)

		// Update log group message to show banned status
		editText := fmt.Sprintf("%s\n\n🚫 *USER BANNED!* (Warn: %d/%d)", callback.Message.Caption, newWarns, constants.MaxWarnings)
		if callback.Message.Photo != nil {
			editCaption := tgbotapi.NewEditMessageCaption(constants.LogGroupID, callback.Message.MessageID, editText)
			editCaption.ParseMode = "Markdown"
			bot.Send(editCaption)
		} else if callback.Message.Video != nil {
			editCaption := tgbotapi.NewEditMessageCaption(constants.LogGroupID, callback.Message.MessageID, editText)
			editCaption.ParseMode = "Markdown"
			bot.Send(editCaption)
		}

		log.Printf("🚫 User %d auto-banned after %d warnings", senderID, newWarns)
	} else {
		// Notify user about warning
		warnMsg := fmt.Sprintf(constants.MsgWarnedNotify, newWarns, constants.MaxWarnings, constants.MaxWarnings)
		p.sendMessage(bot, senderID, warnMsg)

		// Update callback answer
		callbackResponse := tgbotapi.NewCallback(callback.ID, fmt.Sprintf("User %d diberi warning (%d/%d)", senderID, newWarns, constants.MaxWarnings))
		bot.Send(callbackResponse)

		// Update log group message to show warn count
		editText := fmt.Sprintf("%s\n\n⚠️ *WARNED!* (Warn: %d/%d)", callback.Message.Caption, newWarns, constants.MaxWarnings)
		if callback.Message.Photo != nil {
			editCaption := tgbotapi.NewEditMessageCaption(constants.LogGroupID, callback.Message.MessageID, editText)
			editCaption.ParseMode = "Markdown"
			bot.Send(editCaption)
		} else if callback.Message.Video != nil {
			editCaption := tgbotapi.NewEditMessageCaption(constants.LogGroupID, callback.Message.MessageID, editText)
			editCaption.ParseMode = "Markdown"
			bot.Send(editCaption)
		}

		log.Printf("⚠️ User %d warned (%d/%d)", senderID, newWarns, constants.MaxWarnings)
	}

	return nil
}
