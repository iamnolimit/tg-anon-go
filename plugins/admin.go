package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"tg-anon-go/constants"
	"tg-anon-go/databases"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Ad represents a single advertisement
type Ad struct {
	ID        int       `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// AdminPlugin menangani command admin
type AdminPlugin struct {
	BasePlugin
	pendingReset map[int64]bool
}

// NewAdminPlugin membuat instance AdminPlugin baru
func NewAdminPlugin() *AdminPlugin {
	return &AdminPlugin{
		pendingReset: make(map[int64]bool),
	}
}

// Name mengembalikan nama plugin
func (p *AdminPlugin) Name() string {
	return "admin"
}

// Commands mengembalikan daftar command yang ditangani
func (p *AdminPlugin) Commands() []string {
	return []string{
		constants.CmdAdmin,
		constants.CmdBroadcast,
		constants.CmdResetDB,
		constants.CmdAddAds,
		constants.CmdDelAds,
		constants.CmdListAds,
		constants.CmdToggleAds,
		constants.CmdStats,
		constants.CmdBan,
		constants.CmdUnban,
		constants.CmdEnv,
		constants.CmdUpdate,
		"confirmreset",
	}
}

// isOwner mengecek apakah user adalah owner
func (p *AdminPlugin) isOwner(userID int64) bool {
	for _, ownerID := range constants.OwnerIDs {
		if userID == ownerID {
			return true
		}
	}
	return false
}

// HandleCommand menangani command admin
func (p *AdminPlugin) HandleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, command string) error {
	ctx := context.Background()
	chatID := message.Chat.ID
	userID := message.From.ID

	// Check if user is owner
	if !p.isOwner(userID) {
		return p.sendMessage(bot, chatID, constants.MsgAdminOnly)
	}
	switch command {
	case constants.CmdAdmin:
		return p.handleAdminPanel(ctx, bot, chatID)
	case constants.CmdStats:
		return p.handleStats(ctx, bot, chatID)
	case constants.CmdEnv:
		return p.handleEnv(bot, chatID)
	case constants.CmdBroadcast:
		return p.handleBroadcast(ctx, bot, chatID, message)
	case constants.CmdResetDB:
		return p.handleResetDBRequest(bot, chatID, userID)
	case "confirmreset":
		return p.handleResetDBConfirm(ctx, bot, chatID, userID)
	case constants.CmdAddAds:
		return p.handleAddAds(ctx, bot, chatID, message)
	case constants.CmdDelAds:
		return p.handleDelAds(ctx, bot, chatID, message)
	case constants.CmdListAds:
		return p.handleListAds(ctx, bot, chatID)
	case constants.CmdToggleAds:
		return p.handleToggleAds(ctx, bot, chatID)
	case constants.CmdBan:
		return p.handleBan(ctx, bot, chatID, message)
	case constants.CmdUnban:
		return p.handleUnban(ctx, bot, chatID, message)
	case constants.CmdUpdate:
		return p.handleUpdate(ctx, bot, chatID)
	}

	return nil
}

// handleAdminPanel menampilkan panel admin
func (p *AdminPlugin) handleAdminPanel(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error {
	totalUsers, activeChats, _ := databases.GetUserStats(ctx)
	totalMessages, _ := databases.GetGlobalVarInt(ctx, constants.VarGlobalTotalMessages)

	msg := fmt.Sprintf(constants.MsgAdminPanel, totalUsers, activeChats, totalMessages)
	return p.sendMessage(bot, chatID, msg)
}

// handleStats menampilkan statistik
func (p *AdminPlugin) handleStats(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error {
	totalUsers, activeChats, _ := databases.GetUserStats(ctx)
	totalMessages, _ := databases.GetGlobalVarInt(ctx, constants.VarGlobalTotalMessages)

	// Count searching users
	searchingCount, _ := p.countSearchingUsers(ctx)

	// Ads info
	adsEnabled, _ := databases.GetGlobalVarBool(ctx, constants.VarGlobalAdsEnabled)
	adsEnabledStr := "Tidak"
	if adsEnabled {
		adsEnabledStr = "Ya"
	}

	ads, _ := p.getAds(ctx)

	msg := fmt.Sprintf(constants.MsgStatsInfo, totalUsers, activeChats, searchingCount, totalMessages, adsEnabledStr, len(ads))
	return p.sendMessage(bot, chatID, msg)
}

// countSearchingUsers menghitung jumlah user yang sedang searching
func (p *AdminPlugin) countSearchingUsers(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM vars WHERE var_key = $1 AND var_value = $2`
	var count int
	err := databases.DB.QueryRow(ctx, query, constants.VarStatus, constants.StatusSearching).Scan(&count)
	return count, err
}

// handleBroadcast mengirim broadcast ke semua user
func (p *AdminPlugin) handleBroadcast(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, message *tgbotapi.Message) error {
	// Get broadcast message
	args := strings.TrimPrefix(message.Text, "/"+constants.CmdBroadcast)
	args = strings.TrimSpace(args)

	if args == "" {
		return p.sendMessage(bot, chatID, "❌ Format: /broadcast <pesan>")
	}

	// Get all registered users
	users, err := p.getAllRegisteredUsers(ctx)
	if err != nil {
		return p.sendMessage(bot, chatID, constants.MsgError)
	}

	p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgBroadcastStart, len(users)))

	// Send broadcast
	success := 0
	failed := 0
	broadcastMsg := "📢 *Broadcast dari Admin:*\n\n" + args

	for _, userID := range users {
		msg := tgbotapi.NewMessage(userID, broadcastMsg)
		msg.ParseMode = "Markdown"
		_, err := bot.Send(msg)
		if err != nil {
			failed++
		} else {
			success++
		}
		// Rate limiting
		time.Sleep(50 * time.Millisecond)
	}

	return p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgBroadcastDone, success, failed))
}

// getAllRegisteredUsers mengambil semua user ID yang terdaftar
func (p *AdminPlugin) getAllRegisteredUsers(ctx context.Context) ([]int64, error) {
	query := `SELECT telegram_id FROM users`
	rows, err := databases.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		users = append(users, userID)
	}
	return users, nil
}

// handleResetDBRequest meminta konfirmasi reset database
func (p *AdminPlugin) handleResetDBRequest(bot *tgbotapi.BotAPI, chatID int64, userID int64) error {
	p.pendingReset[userID] = true
	return p.sendMessage(bot, chatID, constants.MsgResetDBConfirm)
}

// handleResetDBConfirm melakukan reset database setelah konfirmasi
func (p *AdminPlugin) handleResetDBConfirm(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, userID int64) error {
	if !p.pendingReset[userID] {
		return p.sendMessage(bot, chatID, "❌ Tidak ada permintaan reset yang pending. Gunakan /resetdb terlebih dahulu.")
	}

	delete(p.pendingReset, userID)

	// Reset tables
	queries := []string{
		"TRUNCATE TABLE messages CASCADE",
		"TRUNCATE TABLE chat_sessions CASCADE",
		"TRUNCATE TABLE vars CASCADE",
		"UPDATE users SET status = 'idle', partner_id = NULL",
	}

	for _, query := range queries {
		if _, err := databases.DB.Exec(ctx, query); err != nil {
			log.Printf("Error executing reset query: %v", err)
		}
	}

	return p.sendMessage(bot, chatID, constants.MsgResetDBSuccess)
}

// handleAddAds menambahkan ads baru
func (p *AdminPlugin) handleAddAds(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, message *tgbotapi.Message) error {
	args := strings.TrimPrefix(message.Text, "/"+constants.CmdAddAds)
	args = strings.TrimSpace(args)

	if args == "" {
		return p.sendMessage(bot, chatID, "❌ Format: /addads <pesan iklan>")
	}

	// Get existing ads
	ads, _ := p.getAds(ctx)

	// Generate new ID
	newID := 1
	if len(ads) > 0 {
		newID = ads[len(ads)-1].ID + 1
	}

	// Add new ad
	newAd := Ad{
		ID:        newID,
		Message:   args,
		CreatedAt: time.Now(),
	}
	ads = append(ads, newAd)

	// Save ads
	if err := p.saveAds(ctx, ads); err != nil {
		return p.sendMessage(bot, chatID, constants.MsgError)
	}

	return p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgAdsAdded, newID))
}

// handleDelAds menghapus ads
func (p *AdminPlugin) handleDelAds(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, message *tgbotapi.Message) error {
	args := strings.TrimPrefix(message.Text, "/"+constants.CmdDelAds)
	args = strings.TrimSpace(args)

	adID, err := strconv.Atoi(args)
	if err != nil {
		return p.sendMessage(bot, chatID, "❌ Format: /delads <id>")
	}

	ads, _ := p.getAds(ctx)

	// Find and remove ad
	found := false
	newAds := make([]Ad, 0)
	for _, ad := range ads {
		if ad.ID == adID {
			found = true
			continue
		}
		newAds = append(newAds, ad)
	}

	if !found {
		return p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgAdsNotFound, adID))
	}

	// Save ads
	if err := p.saveAds(ctx, newAds); err != nil {
		return p.sendMessage(bot, chatID, constants.MsgError)
	}

	return p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgAdsDeleted, adID))
}

// handleListAds menampilkan daftar ads
func (p *AdminPlugin) handleListAds(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error {
	ads, _ := p.getAds(ctx)

	if len(ads) == 0 {
		return p.sendMessage(bot, chatID, constants.MsgAdsEmpty)
	}

	var sb strings.Builder
	for _, ad := range ads {
		sb.WriteString(fmt.Sprintf("*ID %d:*\n%s\n\n", ad.ID, ad.Message))
	}

	return p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgAdsList, sb.String()))
}

// handleToggleAds enable/disable ads
func (p *AdminPlugin) handleToggleAds(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error {
	enabled, _ := databases.GetGlobalVarBool(ctx, constants.VarGlobalAdsEnabled)

	// Toggle
	newEnabled := !enabled
	databases.SetGlobalVar(ctx, constants.VarGlobalAdsEnabled, newEnabled)

	status := "Disabled"
	if newEnabled {
		status = "Enabled"
	}

	return p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgAdsToggled, status))
}

// handleBan ban user
func (p *AdminPlugin) handleBan(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, message *tgbotapi.Message) error {
	args := strings.TrimPrefix(message.Text, "/"+constants.CmdBan)
	args = strings.TrimSpace(args)

	userID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		return p.sendMessage(bot, chatID, constants.MsgInvalidUserID)
	}

	databases.BanUser(ctx, userID)
	return p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgUserBanned, userID))
}

// handleUnban unban user
func (p *AdminPlugin) handleUnban(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, message *tgbotapi.Message) error {
	args := strings.TrimPrefix(message.Text, "/"+constants.CmdUnban)
	args = strings.TrimSpace(args)

	userID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		return p.sendMessage(bot, chatID, constants.MsgInvalidUserID)
	}

	databases.UnbanUser(ctx, userID)
	return p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgUserUnbanned, userID))
}

// getAds mengambil semua ads dari database
func (p *AdminPlugin) getAds(ctx context.Context) ([]Ad, error) {
	adsJSON, err := databases.GetGlobalVar(ctx, constants.VarGlobalAds)
	if err != nil || adsJSON == "" {
		return []Ad{}, nil
	}

	var ads []Ad
	if err := json.Unmarshal([]byte(adsJSON), &ads); err != nil {
		return []Ad{}, nil
	}

	return ads, nil
}

// saveAds menyimpan ads ke database
func (p *AdminPlugin) saveAds(ctx context.Context, ads []Ad) error {
	adsJSON, err := json.Marshal(ads)
	if err != nil {
		return err
	}
	return databases.SetGlobalVar(ctx, constants.VarGlobalAds, string(adsJSON))
}

// GetRandomAd mengambil ads secara random
func (p *AdminPlugin) GetRandomAd(ctx context.Context) (string, bool) {
	enabled, _ := databases.GetGlobalVarBool(ctx, constants.VarGlobalAdsEnabled)
	if !enabled {
		return "", false
	}

	ads, _ := p.getAds(ctx)
	if len(ads) == 0 {
		return "", false
	}

	// Random select
	ad := ads[rand.Intn(len(ads))]
	return constants.MsgAdsPrefix + ad.Message, true
}

// handleEnv menampilkan environment variables
func (p *AdminPlugin) handleEnv(bot *tgbotapi.BotAPI, chatID int64) error {
	// Format owner IDs
	ownerIDsStr := make([]string, len(constants.OwnerIDs))
	for i, id := range constants.OwnerIDs {
		ownerIDsStr[i] = strconv.FormatInt(id, 10)
	}
	ownerIDsFormatted := strings.Join(ownerIDsStr, ", ")

	msg := fmt.Sprintf(constants.MsgEnvInfo,
		constants.BotOwnerURL,
		constants.BotChannelURL,
		constants.BotSupportURL,
		constants.LogGroupID,
		ownerIDsFormatted,
		constants.MaxWarnings,
		constants.AdsIntervalMessages,
	)

	return p.sendMessage(bot, chatID, msg)
}

// handleUpdate updates the bot by pulling latest code and rebuilding
func (p *AdminPlugin) handleUpdate(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error {
	// Send start message
	p.sendMessage(bot, chatID, constants.MsgUpdateStart)

	// Step 1: Git pull
	p.sendMessage(bot, chatID, constants.MsgUpdatePulling)
	gitCmd := exec.Command("git", "pull", "origin", "main")
	gitOutput, err := gitCmd.CombinedOutput()
	if err != nil {
		return p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgUpdateFailed, fmt.Sprintf("Git pull error: %v\n%s", err, string(gitOutput))))
	}

	log.Printf("Git pull output: %s", string(gitOutput))

	// Step 2: Build
	p.sendMessage(bot, chatID, constants.MsgUpdateBuilding)
	buildCmd := exec.Command("go", "build", "-o", "tg-anon-go.exe")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		return p.sendMessage(bot, chatID, fmt.Sprintf(constants.MsgUpdateFailed, fmt.Sprintf("Build error: %v\n%s", err, string(buildOutput))))
	}

	log.Printf("Build output: %s", string(buildOutput))

	// Step 3: Success message and restart
	p.sendMessage(bot, chatID, constants.MsgUpdateSuccess)

	// Wait 3 seconds then exit (deployment platform will restart)
	go func() {
		time.Sleep(3 * time.Second)
		log.Println("Exiting for update restart...")
		os.Exit(0)
	}()

	return nil
}

func (p *AdminPlugin) sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := bot.Send(msg)
	return err
}
