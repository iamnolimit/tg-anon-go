package constants

import (
	"os"
	"strconv"
	"strings"
)

// GetEnv helper function to get env with default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt64 helper function to get env as int64 with default value
func GetEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetEnvInt helper function to get env as int with default value
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetEnvInt64Slice helper function to get env as []int64 (comma-separated)
func GetEnvInt64Slice(key string, defaultValue []int64) []int64 {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]int64, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if intVal, err := strconv.ParseInt(part, 10, 64); err == nil {
				result = append(result, intVal)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

// Bot Info URLs - dari environment variables
var (
	BotOwnerURL   = GetEnv("BOT_OWNER_URL", "https://t.me/ursweetbae")
	BotChannelURL = GetEnv("BOT_CHANNEL_URL", "https://t.me/baecorner")
	BotSupportURL = GetEnv("BOT_SUPPORT_URL", "https://t.me/tgnolimitchat")
)

// Redis URL - dari environment variable with default
var DefaultRedisURL = "redis://default:5NQHBzWOhwHrczAy8SfFtqCCoPcHVTzn@redis-12448.crce194.ap-seast-1-1.ec2.cloud.redislabs.com:12448"

// Log Group ID - dari environment variable
var LogGroupID = GetEnvInt64("LOG_GROUP_ID", -1002339919418)

// Owner IDs - dari environment variable (comma-separated)
var OwnerIDs = GetEnvInt64Slice("OWNER_IDS", []int64{1259894923})

// Warn Settings - dari environment variable
var MaxWarnings = GetEnvInt("MAX_WARNINGS", 3)

// Ads Settings - dari environment variable
var AdsIntervalMessages = GetEnvInt("ADS_INTERVAL_MESSAGES", 30)

// Variable Keys untuk user
const (
	// User Status & Partner
	VarStatus    = "status"     // Status user: idle, searching, chatting
	VarPartnerID = "partner_id" // ID partner yang sedang chat

	// User Profile
	VarName      = "name"      // Nama user
	VarAge       = "age"       // Umur user
	VarLocation  = "location"  // Nama lokasi user
	VarLatitude  = "latitude"  // Koordinat latitude
	VarLongitude = "longitude" // Koordinat longitude
	// Registration State
	VarRegState  = "reg_state"  // State registrasi: none, ask_name, ask_age, ask_location, done
	VarEditState = "edit_state" // State edit profil: none, edit_name, edit_age, edit_gender, edit_location

	// Search Preferences
	VarSearchMode   = "search_mode"   // Mode pencarian: random, nearby
	VarGender       = "gender"        // Jenis kelamin user
	VarSearchGender = "search_gender" // Preferensi gender yang dicari

	// Statistics
	VarTotalChats    = "total_chats"    // Total chat yang dilakukan
	VarTotalMessages = "total_messages" // Total pesan yang dikirim
	VarLastActive    = "last_active"    // Timestamp terakhir aktif

	// Session
	VarSessionID    = "session_id"    // ID sesi chat aktif
	VarSessionStart = "session_start" // Waktu mulai sesi

	// Flags
	VarIsBanned    = "is_banned"    // Apakah user dibanned
	VarIsVerified  = "is_verified"  // Apakah user terverifikasi
	VarIsPremium   = "is_premium"   // Apakah user premium
	VarIsRegistered = "is_registered" // Apakah user sudah registrasi
	VarWarnCount   = "warn_count"   // Jumlah warn user

	// Settings
	VarNotifications = "notifications" // Notifikasi enabled/disabled
	VarLanguage      = "language"      // Bahasa preferensi
)

// Global Variable Keys (userID = 0)
const (
	VarGlobalTotalUsers    = "global_total_users"    // Total user terdaftar
	VarGlobalActiveChats   = "global_active_chats"   // Total chat aktif
	VarGlobalTotalMessages = "global_total_messages" // Total pesan
	VarGlobalBotStatus     = "global_bot_status"     // Status bot (maintenance, active)
	VarGlobalAds           = "global_ads"            // Daftar ads dalam JSON
	VarGlobalAdsEnabled    = "global_ads_enabled"    // Ads enabled/disabled
)

// Gender Values
const (
	GenderMale   = "Pria"
	GenderFemale = "Wanita"
	GenderOther  = "Lainnya"
	GenderAny    = "any"
)

// Search Mode Values
const (
	SearchModeRandom = "random"
	SearchModeNearby = "nearby"
)

// Registration States
const (
	RegStateNone        = "none"
	RegStateAskName     = "ask_name"
	RegStateAskAge      = "ask_age"
	RegStateAskGender   = "ask_gender"
	RegStateAskLocation = "ask_location"
	RegStateDone        = "done"
)

// Edit Profile States
const (
	EditStateNone     = "none"
	EditStateName     = "edit_name"
	EditStateAge      = "edit_age"
	EditStateGender   = "edit_gender"
	EditStateLocation = "edit_location"
)
