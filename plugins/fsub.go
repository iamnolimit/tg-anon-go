package plugins

import (
	"context"
	"log"
	"strings"

	"tg-anon-go/constants"
	"tg-anon-go/databases"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CheckFsub checks if user is member of required channel
// Returns (allowed bool, channelUsername string)
func CheckFsub(ctx context.Context, bot *tgbotapi.BotAPI, userID int64) (bool, string) {
	// Check if fsub enabled
	enabled, _ := databases.GetGlobalVarBool(ctx, constants.VarGlobalFsubEnabled)
	if !enabled {
		return true, "" // No fsub, allow access
	}

	// Get channel
	channel, _ := databases.GetGlobalVar(ctx, constants.VarGlobalFsubChannel)
	if channel == "" {
		return true, "" // No channel set
	}

	// Check membership using SuperGroupUsername which accepts both @username and numeric -100xxx format
	chatMember, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			SuperGroupUsername: channel,
			UserID:             userID,
		},
	})

	if err != nil {
		log.Printf("Error checking channel membership for user %d: %v", userID, err)
		return false, channel
	}

	// Check if user is member/admin/creator
	status := chatMember.Status
	if status == "member" || status == "administrator" || status == "creator" {
		return true, ""
	}

	return false, channel
}

// SendFsubPrompt sends join channel prompt with buttons
func SendFsubPrompt(bot *tgbotapi.BotAPI, chatID int64, channelUsername string) error {
	msg := tgbotapi.NewMessage(chatID, constants.MsgFsubRequired)
	msg.ParseMode = "Markdown"

	// Create URL for channel
	channelURL := channelUsername
	if strings.HasPrefix(channelUsername, "@") {
		// Username format: @channelname
		channelURL = "https://t.me/" + strings.TrimPrefix(channelUsername, "@")
	} else if strings.HasPrefix(channelUsername, "-100") {
		// Private channel ID format: -100123456789
		// Can't create direct link for private channels
		channelURL = ""
	}

	var keyboard tgbotapi.InlineKeyboardMarkup
	if channelURL != "" {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("📢 Join Channel", channelURL),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Sudah Join", "fsub_verify"),
			),
		)
	} else {
		// For private channels, only show verify button
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Sudah Join", "fsub_verify"),
			),
		)
	}
	msg.ReplyMarkup = keyboard

	_, err := bot.Send(msg)
	return err
}
