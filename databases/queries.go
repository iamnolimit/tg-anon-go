package databases

import (
	"context"
	"fmt"
	"time"

	"tg-anon-go/constants"
)

// User represents a user in the database
type User struct {
	ID         int64
	TelegramID int64
	Username   string
	FirstName  string
	Status     string
	PartnerID  *int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ChatSession represents a chat session
type ChatSession struct {
	ID        int64
	User1ID   int64
	User2ID   int64
	StartedAt time.Time
	EndedAt   *time.Time
	IsActive  bool
}

// CreateOrUpdateUser membuat atau memperbarui user
func CreateOrUpdateUser(ctx context.Context, telegramID int64, username, firstName string) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name, status, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (telegram_id)
		DO UPDATE SET username = $2, first_name = $3, updated_at = $5
	`
	_, err := DB.Exec(ctx, query, telegramID, username, firstName, constants.StatusIdle, time.Now())
	return err
}

// GetUserByTelegramID mengambil user berdasarkan telegram ID
func GetUserByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, status, partner_id, created_at, updated_at
		FROM users WHERE telegram_id = $1
	`
	user := &User{}
	err := DB.QueryRow(ctx, query, telegramID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.Status, &user.PartnerID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserStatus memperbarui status user
func UpdateUserStatus(ctx context.Context, telegramID int64, status string, partnerID *int64) error {
	query := `
		UPDATE users SET status = $1, partner_id = $2, updated_at = $3
		WHERE telegram_id = $4
	`
	_, err := DB.Exec(ctx, query, status, partnerID, time.Now(), telegramID)
	return err
}

// FindSearchingUser mencari user yang sedang mencari partner (selain diri sendiri)
func FindSearchingUser(ctx context.Context, excludeTelegramID int64) (*User, error) {
	// Cari dari tabel vars dimana status = searching
	query := `
		SELECT u.id, u.telegram_id, u.username, u.first_name, 
			   COALESCE(v.var_value, 'idle') as status, 
			   u.partner_id, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN vars v ON u.telegram_id = v.user_id AND v.var_key = $1
		WHERE v.var_value = $2 AND u.telegram_id != $3
		ORDER BY v.updated_at ASC
		LIMIT 1
	`
	user := &User{}
	err := DB.QueryRow(ctx, query, constants.VarStatus, constants.StatusSearching, excludeTelegramID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.Status, &user.PartnerID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateChatSession membuat sesi chat baru
func CreateChatSession(ctx context.Context, user1ID, user2ID int64) (int64, error) {
	query := `
		INSERT INTO chat_sessions (user1_id, user2_id, started_at, is_active)
		VALUES ($1, $2, $3, true)
		RETURNING id
	`
	var sessionID int64
	err := DB.QueryRow(ctx, query, user1ID, user2ID, time.Now()).Scan(&sessionID)
	return sessionID, err
}

// EndChatSession mengakhiri sesi chat
func EndChatSession(ctx context.Context, user1ID, user2ID int64) error {
	query := `
		UPDATE chat_sessions 
		SET is_active = false, ended_at = $1
		WHERE ((user1_id = $2 AND user2_id = $3) OR (user1_id = $3 AND user2_id = $2))
		AND is_active = true
	`
	_, err := DB.Exec(ctx, query, time.Now(), user1ID, user2ID)
	return err
}

// SaveMessage menyimpan pesan ke database
func SaveMessage(ctx context.Context, sessionID int64, senderID, receiverID int64, msgType, content string) error {
	query := `
		INSERT INTO messages (session_id, sender_id, receiver_id, message_type, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := DB.Exec(ctx, query, sessionID, senderID, receiverID, msgType, content, time.Now())
	return err
}

// GetActiveSessionID mengambil ID sesi aktif untuk user
func GetActiveSessionID(ctx context.Context, userID int64) (int64, error) {
	query := `
		SELECT id FROM chat_sessions
		WHERE (user1_id = $1 OR user2_id = $1) AND is_active = true
		ORDER BY started_at DESC
		LIMIT 1
	`
	var sessionID int64
	err := DB.QueryRow(ctx, query, userID).Scan(&sessionID)
	return sessionID, err
}

// GetUserStats mengambil statistik user
func GetUserStats(ctx context.Context) (totalUsers int64, activeChats int64, err error) {
	err = DB.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return
	}
	err = DB.QueryRow(ctx, "SELECT COUNT(*) FROM chat_sessions WHERE is_active = true").Scan(&activeChats)
	return
}

// GetOldActiveSessions mengambil sesi yang sudah aktif lebih lama dari durasi tertentu
func GetOldActiveSessions(ctx context.Context, maxAge time.Duration) ([]ChatSession, error) {
	query := `
		SELECT id, user1_id, user2_id, started_at, ended_at, is_active
		FROM chat_sessions
		WHERE is_active = true AND started_at < $1
		ORDER BY started_at ASC
	`
	cutoffTime := time.Now().Add(-maxAge)
	rows, err := DB.Query(ctx, query, cutoffTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []ChatSession
	for rows.Next() {
		var session ChatSession
		if err := rows.Scan(&session.ID, &session.User1ID, &session.User2ID,
			&session.StartedAt, &session.EndedAt, &session.IsActive); err != nil {
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// ============================================================
// HELPER FUNCTIONS MENGGUNAKAN VAR SYSTEM
// ============================================================

// SetUserStatus mengatur status user menggunakan var system
func SetUserStatus(ctx context.Context, userID int64, status string) error {
	return SetVar(ctx, userID, constants.VarStatus, status)
}

// GetUserStatus mengambil status user menggunakan var system
func GetUserStatus(ctx context.Context, userID int64) (string, error) {
	status, err := GetVar(ctx, userID, constants.VarStatus)
	if err != nil {
		return "", err
	}
	if status == "" {
		return constants.StatusIdle, nil
	}
	return status, nil
}

// SetUserPartner mengatur partner ID untuk user
func SetUserPartner(ctx context.Context, userID int64, partnerID int64) error {
	return SetVar(ctx, userID, constants.VarPartnerID, partnerID)
}

// GetUserPartner mengambil partner ID user
func GetUserPartner(ctx context.Context, userID int64) (int64, error) {
	return GetVarInt64(ctx, userID, constants.VarPartnerID)
}

// ClearUserPartner menghapus partner ID user
func ClearUserPartner(ctx context.Context, userID int64) error {
	return DeleteVar(ctx, userID, constants.VarPartnerID)
}

// SetUserSessionID mengatur session ID untuk user
func SetUserSessionID(ctx context.Context, userID int64, sessionID int64) error {
	return SetVar(ctx, userID, constants.VarSessionID, sessionID)
}

// GetUserSessionID mengambil session ID user
func GetUserSessionID(ctx context.Context, userID int64) (int64, error) {
	return GetVarInt64(ctx, userID, constants.VarSessionID)
}

// IncrementUserTotalChats menambah total chat user
func IncrementUserTotalChats(ctx context.Context, userID int64) error {
	current, _ := GetVarInt(ctx, userID, constants.VarTotalChats)
	return SetVar(ctx, userID, constants.VarTotalChats, current+1)
}

// IncrementUserTotalMessages menambah total pesan user
func IncrementUserTotalMessages(ctx context.Context, userID int64) error {
	current, _ := GetVarInt(ctx, userID, constants.VarTotalMessages)
	return SetVar(ctx, userID, constants.VarTotalMessages, current+1)
}

// UpdateLastActive mengupdate waktu terakhir aktif user
func UpdateLastActive(ctx context.Context, userID int64) error {
	return SetVar(ctx, userID, constants.VarLastActive, time.Now().Unix())
}

// IsUserBanned mengecek apakah user dibanned
func IsUserBanned(ctx context.Context, userID int64) (bool, error) {
	return GetVarBool(ctx, userID, constants.VarIsBanned)
}

// BanUser mem-ban user
func BanUser(ctx context.Context, userID int64) error {
	return SetVar(ctx, userID, constants.VarIsBanned, true)
}

// UnbanUser meng-unban user
func UnbanUser(ctx context.Context, userID int64) error {
	return SetVar(ctx, userID, constants.VarIsBanned, false)
}

// ConnectUsers menghubungkan dua user untuk chat
func ConnectUsers(ctx context.Context, user1ID, user2ID int64) (int64, error) {
	// Create session
	sessionID, err := CreateChatSession(ctx, user1ID, user2ID)
	if err != nil {
		return 0, err
	}

	// Set status dan partner untuk kedua user
	if err := SetUserStatus(ctx, user1ID, constants.StatusChatting); err != nil {
		return 0, err
	}
	if err := SetUserStatus(ctx, user2ID, constants.StatusChatting); err != nil {
		return 0, err
	}
	if err := SetUserPartner(ctx, user1ID, user2ID); err != nil {
		return 0, err
	}
	if err := SetUserPartner(ctx, user2ID, user1ID); err != nil {
		return 0, err
	}
	if err := SetUserSessionID(ctx, user1ID, sessionID); err != nil {
		return 0, err
	}
	if err := SetUserSessionID(ctx, user2ID, sessionID); err != nil {
		return 0, err
	}

	// Increment total chats
	IncrementUserTotalChats(ctx, user1ID)
	IncrementUserTotalChats(ctx, user2ID)

	return sessionID, nil
}

// DisconnectUsers memutuskan koneksi chat antara dua user
func DisconnectUsers(ctx context.Context, user1ID, user2ID int64) error {
	// Get session ID
	sessionID, _ := GetUserSessionID(ctx, user1ID)

	// End session di database
	if err := EndChatSession(ctx, user1ID, user2ID); err != nil {
		return err
	}

	// Reset status dan partner untuk kedua user
	SetUserStatus(ctx, user1ID, constants.StatusIdle)
	SetUserStatus(ctx, user2ID, constants.StatusIdle)
	ClearUserPartner(ctx, user1ID)
	ClearUserPartner(ctx, user2ID)
	DeleteVar(ctx, user1ID, constants.VarSessionID)
	DeleteVar(ctx, user2ID, constants.VarSessionID)
	_ = sessionID // untuk menghindari unused variable warning
	return nil
}

// FindAndConnectPartner mencari dan menghubungkan dengan partner yang sedang searching
func FindAndConnectPartner(ctx context.Context, userID int64) (*User, int64, error) {
	// Cari user yang sedang searching
	partner, err := FindSearchingUser(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	// Hubungkan kedua user
	sessionID, err := ConnectUsers(ctx, userID, partner.TelegramID)
	if err != nil {
		return nil, 0, err
	}

	return partner, sessionID, nil
}

// SearchingUserWithLocation represents a searching user with location data
type SearchingUserWithLocation struct {
	User     *User
	Distance float64
}

// FindNearbySearchingUser mencari partner terdekat yang sedang searching
func FindNearbySearchingUser(ctx context.Context, userID int64, maxDistanceKm float64) (*User, float64, error) {
	// Get user's location
	userLat, userLon, err := GetUserLocation(ctx, userID)
	if err != nil || (userLat == 0 && userLon == 0) {
		return nil, 0, err
	}

	// Get all searching users
	query := `
		SELECT DISTINCT v.user_id 
		FROM vars v 
		WHERE v.var_key = $1 AND v.var_value = $2 AND v.user_id != $3
	`
	rows, err := DB.Query(ctx, query, constants.VarStatus, constants.StatusSearching, userID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var nearestUser *User
	var nearestDistance float64 = -1

	for rows.Next() {
		var partnerID int64
		if err := rows.Scan(&partnerID); err != nil {
			continue
		}

		// Get partner's location
		partnerLat, partnerLon, err := GetUserLocation(ctx, partnerID)
		if err != nil || (partnerLat == 0 && partnerLon == 0) {
			continue
		}

		// Calculate distance
		distance := CalculateDistance(userLat, userLon, partnerLat, partnerLon)

		// Check if within max distance and is nearest
		if distance <= maxDistanceKm && (nearestDistance < 0 || distance < nearestDistance) {
			user, err := GetUserByTelegramID(ctx, partnerID)
			if err == nil {
				nearestUser = user
				nearestDistance = distance
			}
		}
	}

	if nearestUser == nil {
		return nil, 0, fmt.Errorf("no nearby partner found")
	}

	return nearestUser, nearestDistance, nil
}

// FindAndConnectNearbyPartner mencari dan menghubungkan dengan partner terdekat
func FindAndConnectNearbyPartner(ctx context.Context, userID int64, maxDistanceKm float64) (*User, int64, float64, error) {
	// Cari user terdekat yang sedang searching
	partner, distance, err := FindNearbySearchingUser(ctx, userID, maxDistanceKm)
	if err != nil {
		return nil, 0, 0, err
	}

	// Hubungkan kedua user
	sessionID, err := ConnectUsers(ctx, userID, partner.TelegramID)
	if err != nil {
		return nil, 0, 0, err
	}

	return partner, sessionID, distance, nil
}
