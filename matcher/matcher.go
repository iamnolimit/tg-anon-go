package matcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"tg-anon-go/constants"
	"tg-anon-go/databases"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis channels
	ChannelSearchRandom = "search:random"
	ChannelSearchNearby = "search:nearby"

	// Redis keys
	KeySearchingUsers = "searching:users" // Set of all searching user IDs
	KeyUserData       = "user:%d:data"    // Hash of user search data
	KeyMatchLock      = "match:lock:%d"   // Lock to prevent double matching
	LockExpiration    = 10 * time.Second  // Lock expiration time
	UserDataTTL       = 5 * time.Minute   // User data expiration
)

// SearchRequest represents a search request from a user
type SearchRequest struct {
	UserID     int64   `json:"user_id"`
	SearchMode string  `json:"search_mode"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
	Timestamp  int64   `json:"timestamp"`
}

// Matcher menangani realtime matching menggunakan Redis Pub/Sub
type Matcher struct {
	bot       *tgbotapi.BotAPI
	rdb       *redis.Client
	stopChan  chan struct{}
	wg        sync.WaitGroup
	running   bool
	mu        sync.Mutex
	ctx       context.Context
	cancelCtx context.CancelFunc
}

// NewMatcher membuat instance Matcher baru dengan Redis
func NewMatcher(bot *tgbotapi.BotAPI, redisURL string) (*Matcher, error) {
	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Create Redis client
	rdb := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✅ Connected to Redis successfully")

	ctx, cancel := context.WithCancel(context.Background())

	return &Matcher{
		bot:       bot,
		rdb:       rdb,
		stopChan:  make(chan struct{}),
		ctx:       ctx,
		cancelCtx: cancel,
	}, nil
}

// Start memulai Redis Pub/Sub listeners
func (m *Matcher) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	// Start random matcher
	m.wg.Add(1)
	go m.listenRandom()

	// Start nearby matcher
	m.wg.Add(1)
	go m.listenNearby()

	// Start cleanup worker (remove stale users)
	m.wg.Add(1)
	go m.cleanupWorker()

	// Start periodic matching worker to fix race conditions
	m.wg.Add(1)
	go m.matchingWorker()

	log.Println("🚀 Redis Matcher started (Random + Nearby + Cleanup + Periodic Matcher)")
}

// Stop menghentikan matcher
func (m *Matcher) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	m.mu.Unlock()

	m.cancelCtx()
	close(m.stopChan)
	m.wg.Wait()

	if err := m.rdb.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("⏹️ Redis Matcher stopped")
}

// PublishSearch publishes a search request to Redis
func (m *Matcher) PublishSearch(ctx context.Context, userID int64, searchMode string, lat, lon float64) error {
	req := SearchRequest{
		UserID:     userID,
		SearchMode: searchMode,
		Latitude:   lat,
		Longitude:  lon,
		Timestamp:  time.Now().Unix(),
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal search request: %w", err)
	}

	// Add to searching users set
	if err := m.rdb.SAdd(ctx, KeySearchingUsers, userID).Err(); err != nil {
		return fmt.Errorf("failed to add user to searching set: %w", err)
	}

	// Store user data with TTL
	userKey := fmt.Sprintf(KeyUserData, userID)
	if err := m.rdb.Set(ctx, userKey, data, UserDataTTL).Err(); err != nil {
		return fmt.Errorf("failed to store user data: %w", err)
	}

	// Publish to appropriate channel
	channel := ChannelSearchRandom
	if searchMode == constants.SearchModeNearby {
		channel = ChannelSearchNearby
	}

	if err := m.rdb.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish search request: %w", err)
	}

	log.Printf("📡 Published search request: User %d, Mode: %s", userID, searchMode)
	return nil
}

// RemoveSearchingUser removes a user from searching state
func (m *Matcher) RemoveSearchingUser(ctx context.Context, userID int64) error {
	// Remove from set
	if err := m.rdb.SRem(ctx, KeySearchingUsers, userID).Err(); err != nil {
		return err
	}

	// Delete user data
	userKey := fmt.Sprintf(KeyUserData, userID)
	if err := m.rdb.Del(ctx, userKey).Err(); err != nil {
		return err
	}

	log.Printf("🗑️ Removed user %d from searching", userID)
	return nil
}

// listenRandom listens for random search requests
func (m *Matcher) listenRandom() {
	defer m.wg.Done()

	pubsub := m.rdb.Subscribe(m.ctx, ChannelSearchRandom)
	defer pubsub.Close()

	log.Println("👂 Listening on channel:", ChannelSearchRandom)

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg == nil {
				continue
			}

			var req SearchRequest
			if err := json.Unmarshal([]byte(msg.Payload), &req); err != nil {
				log.Printf("Error unmarshaling random search request: %v", err)
				continue
			}

			m.handleRandomSearch(&req)

		case <-m.stopChan:
			return
		case <-m.ctx.Done():
			return
		}
	}
}

// listenNearby listens for nearby search requests
func (m *Matcher) listenNearby() {
	defer m.wg.Done()

	pubsub := m.rdb.Subscribe(m.ctx, ChannelSearchNearby)
	defer pubsub.Close()

	log.Println("👂 Listening on channel:", ChannelSearchNearby)

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg == nil {
				continue
			}

			var req SearchRequest
			if err := json.Unmarshal([]byte(msg.Payload), &req); err != nil {
				log.Printf("Error unmarshaling nearby search request: %v", err)
				continue
			}

			m.handleNearbySearch(&req)

		case <-m.stopChan:
			return
		case <-m.ctx.Done():
			return
		}
	}
}

// handleRandomSearch handles random matching
func (m *Matcher) handleRandomSearch(req *SearchRequest) {
	ctx := context.Background()

	// Try to acquire lock
	lockKey := fmt.Sprintf(KeyMatchLock, req.UserID)
	locked, err := m.rdb.SetNX(ctx, lockKey, "1", LockExpiration).Result()
	if err != nil || !locked {
		// Already being processed or locked
		return
	}
	defer m.rdb.Del(ctx, lockKey)

	// Check if user still searching in database
	status, _ := databases.GetVar(ctx, req.UserID, constants.VarStatus)
	if status != constants.StatusSearching {
		m.RemoveSearchingUser(ctx, req.UserID)
		return
	}

	// Get all searching users
	members, err := m.rdb.SMembers(ctx, KeySearchingUsers).Result()
	if err != nil {
		log.Printf("Error getting searching users: %v", err)
		return
	}
	// Find a partner
	for _, memberStr := range members {
		var partnerID int64
		fmt.Sscanf(memberStr, "%d", &partnerID)

		if partnerID == req.UserID {
			continue
		}

		// Get partner data
		partnerKey := fmt.Sprintf(KeyUserData, partnerID)
		partnerData, err := m.rdb.Get(ctx, partnerKey).Result()
		if err != nil {
			continue
		}

		var partnerReq SearchRequest
		if err := json.Unmarshal([]byte(partnerData), &partnerReq); err != nil {
			continue
		}

		// Try to lock partner
		partnerLockKey := fmt.Sprintf(KeyMatchLock, partnerID)
		partnerLocked, err := m.rdb.SetNX(ctx, partnerLockKey, "1", LockExpiration).Result()
		if err != nil || !partnerLocked {
			continue
		}

		// Check if partner still searching
		partnerStatus, _ := databases.GetVar(ctx, partnerID, constants.VarStatus)
		if partnerStatus != constants.StatusSearching {
			m.rdb.Del(ctx, partnerLockKey)
			m.RemoveSearchingUser(ctx, partnerID)
			continue
		}

		// Match found! Connect users
		_, err = databases.ConnectUsers(ctx, req.UserID, partnerID)
		if err != nil {
			log.Printf("Error connecting users: %v", err)
			m.rdb.Del(ctx, partnerLockKey)
			continue
		}

		// Remove both from searching
		m.RemoveSearchingUser(ctx, req.UserID)
		m.RemoveSearchingUser(ctx, partnerID)
		m.rdb.Del(ctx, partnerLockKey)

		// Notify both users
		m.notifyMatch(req.UserID, partnerID, constants.SearchModeRandom, 0)

		log.Printf("✅ Random Match: User %d <-> User %d", req.UserID, partnerID)
		return
	}

	log.Printf("⏳ No partner found for user %d yet (random)", req.UserID)
}

// handleNearbySearch handles nearby matching with distance-based priority
func (m *Matcher) handleNearbySearch(req *SearchRequest) {
	ctx := context.Background()

	// Try to acquire lock
	lockKey := fmt.Sprintf(KeyMatchLock, req.UserID)
	locked, err := m.rdb.SetNX(ctx, lockKey, "1", LockExpiration).Result()
	if err != nil || !locked {
		return
	}
	defer m.rdb.Del(ctx, lockKey)

	// Check if user still searching
	status, _ := databases.GetVar(ctx, req.UserID, constants.VarStatus)
	if status != constants.StatusSearching {
		m.RemoveSearchingUser(ctx, req.UserID)
		return
	}

	// Validate user has location
	if req.Latitude == 0 && req.Longitude == 0 {
		log.Printf("User %d has no location, fallback to random", req.UserID)
		req.SearchMode = constants.SearchModeRandom
		m.handleRandomSearch(req)
		return
	}

	// Get all searching users
	members, err := m.rdb.SMembers(ctx, KeySearchingUsers).Result()
	if err != nil {
		log.Printf("Error getting searching users: %v", err)
		return
	}

	// Collect all potential partners with their distances
	type partnerWithDistance struct {
		partnerID int64
		distance  float64
		req       SearchRequest
	}
	var candidates []partnerWithDistance

	for _, memberStr := range members {
		var partnerID int64
		fmt.Sscanf(memberStr, "%d", &partnerID)

		if partnerID == req.UserID {
			continue
		}

		// Get partner data
		partnerKey := fmt.Sprintf(KeyUserData, partnerID)
		partnerData, err := m.rdb.Get(ctx, partnerKey).Result()
		if err != nil {
			continue
		}

		var partnerReq SearchRequest
		if err := json.Unmarshal([]byte(partnerData), &partnerReq); err != nil {
			continue
		}

		// Calculate distance (use max distance if partner has no location)
		var distance float64
		if partnerReq.Latitude != 0 || partnerReq.Longitude != 0 {
			distance = calculateDistance(req.Latitude, req.Longitude, partnerReq.Latitude, partnerReq.Longitude)
		} else {
			distance = 9999 // No location, lowest priority
		}

		candidates = append(candidates, partnerWithDistance{
			partnerID: partnerID,
			distance:  distance,
			req:       partnerReq,
		})
	}

	// Sort candidates by distance (closest first)
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].distance < candidates[i].distance {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Try to match with closest available partner
	for _, candidate := range candidates {
		// Try to lock partner
		partnerLockKey := fmt.Sprintf(KeyMatchLock, candidate.partnerID)
		partnerLocked, err := m.rdb.SetNX(ctx, partnerLockKey, "1", LockExpiration).Result()
		if err != nil || !partnerLocked {
			continue
		}

		// Check if partner still searching
		partnerStatus, _ := databases.GetVar(ctx, candidate.partnerID, constants.VarStatus)
		if partnerStatus != constants.StatusSearching {
			m.rdb.Del(ctx, partnerLockKey)
			m.RemoveSearchingUser(ctx, candidate.partnerID)
			continue
		}

		// Match found! Connect users
		_, err = databases.ConnectUsers(ctx, req.UserID, candidate.partnerID)
		if err != nil {
			log.Printf("Error connecting users: %v", err)
			m.rdb.Del(ctx, partnerLockKey)
			continue
		}

		// Remove both from searching
		m.RemoveSearchingUser(ctx, req.UserID)
		m.RemoveSearchingUser(ctx, candidate.partnerID)
		m.rdb.Del(ctx, partnerLockKey)

		// Notify both users (show distance if nearby)
		if candidate.distance < 9999 {
			m.notifyMatch(req.UserID, candidate.partnerID, constants.SearchModeNearby, candidate.distance)
			log.Printf("✅ Nearby Match: User %d <-> User %d (%.2f km)", req.UserID, candidate.partnerID, candidate.distance)
		} else {
			m.notifyMatch(req.UserID, candidate.partnerID, constants.SearchModeRandom, 0)
			log.Printf("✅ Random Match (fallback): User %d <-> User %d", req.UserID, candidate.partnerID)
		}
		return
	}

	log.Printf("⏳ No partner found for user %d yet (nearby)", req.UserID)
}

// cleanupWorker removes stale searching users periodically
func (m *Matcher) cleanupWorker() {
	defer m.wg.Done()

	// Cleanup every 2 minutes (reduced from 30s to save resources)
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopChan:
			return
		case <-m.ctx.Done():
			return
		}
	}
}

// cleanup removes users who are no longer searching
func (m *Matcher) cleanup() {
	ctx := context.Background()

	members, err := m.rdb.SMembers(ctx, KeySearchingUsers).Result()
	if err != nil {
		return
	}

	for _, memberStr := range members {
		var userID int64
		fmt.Sscanf(memberStr, "%d", &userID)

		// Check database status
		status, _ := databases.GetVar(ctx, userID, constants.VarStatus)
		if status != constants.StatusSearching {
			m.RemoveSearchingUser(ctx, userID)
		}
	}
}

// matchingWorker runs periodically to retry matching stuck users
func (m *Matcher) matchingWorker() {
	defer m.wg.Done()

	// Run matching every 5 seconds to prevent stuck users
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("🔄 Periodic matching worker started (every 5s)")

	for {
		select {
		case <-ticker.C:
			m.retryMatchAllUsers()
		case <-m.stopChan:
			return
		case <-m.ctx.Done():
			return
		}
	}
}

// retryMatchAllUsers tries to match all users currently in the searching set
func (m *Matcher) retryMatchAllUsers() {
	ctx := context.Background()

	// Get all searching users
	members, err := m.rdb.SMembers(ctx, KeySearchingUsers).Result()
	if err != nil {
		return
	}

	// Need at least 2 users to match
	if len(members) < 2 {
		return
	}

	log.Printf("🔄 Retry matching %d searching users...", len(members))

	// Collect valid searching users with their data
	type userWithData struct {
		userID    int64
		searchReq SearchRequest
	}
	var validUsers []userWithData

	for _, memberStr := range members {
		var userID int64
		fmt.Sscanf(memberStr, "%d", &userID)

		// Verify still searching in database
		status, _ := databases.GetVar(ctx, userID, constants.VarStatus)
		if status != constants.StatusSearching {
			m.RemoveSearchingUser(ctx, userID)
			continue
		}

		// Get user search data
		userKey := fmt.Sprintf(KeyUserData, userID)
		userData, err := m.rdb.Get(ctx, userKey).Result()
		if err != nil {
			// No data means TTL expired, refresh it
			searchMode, _ := databases.GetVar(ctx, userID, constants.VarSearchMode)
			if searchMode == "" {
				searchMode = constants.SearchModeRandom
			}
			var lat, lon float64
			if searchMode == constants.SearchModeNearby {
				lat, _ = databases.GetVarFloat64(ctx, userID, constants.VarLatitude)
				lon, _ = databases.GetVarFloat64(ctx, userID, constants.VarLongitude)
			}

			req := SearchRequest{
				UserID:     userID,
				SearchMode: searchMode,
				Latitude:   lat,
				Longitude:  lon,
				Timestamp:  time.Now().Unix(),
			}
			data, _ := json.Marshal(req)
			m.rdb.Set(ctx, userKey, data, UserDataTTL)

			validUsers = append(validUsers, userWithData{userID: userID, searchReq: req})
			continue
		}

		var req SearchRequest
		if err := json.Unmarshal([]byte(userData), &req); err != nil {
			continue
		}
		validUsers = append(validUsers, userWithData{userID: userID, searchReq: req})
	}

	// Try to match users pairwise
	matched := make(map[int64]bool)

	for i := 0; i < len(validUsers); i++ {
		if matched[validUsers[i].userID] {
			continue
		}

		for j := i + 1; j < len(validUsers); j++ {
			if matched[validUsers[j].userID] {
				continue
			}

			user1 := validUsers[i]
			user2 := validUsers[j]

			// Try to lock both users
			lockKey1 := fmt.Sprintf(KeyMatchLock, user1.userID)
			lockKey2 := fmt.Sprintf(KeyMatchLock, user2.userID)

			locked1, err := m.rdb.SetNX(ctx, lockKey1, "1", LockExpiration).Result()
			if err != nil || !locked1 {
				continue
			}

			locked2, err := m.rdb.SetNX(ctx, lockKey2, "1", LockExpiration).Result()
			if err != nil || !locked2 {
				m.rdb.Del(ctx, lockKey1)
				continue
			}

			// Double-check both still searching
			status1, _ := databases.GetVar(ctx, user1.userID, constants.VarStatus)
			status2, _ := databases.GetVar(ctx, user2.userID, constants.VarStatus)

			if status1 != constants.StatusSearching || status2 != constants.StatusSearching {
				m.rdb.Del(ctx, lockKey1)
				m.rdb.Del(ctx, lockKey2)
				if status1 != constants.StatusSearching {
					m.RemoveSearchingUser(ctx, user1.userID)
				}
				if status2 != constants.StatusSearching {
					m.RemoveSearchingUser(ctx, user2.userID)
				}
				continue
			}

			// Match found! Connect users
			_, err = databases.ConnectUsers(ctx, user1.userID, user2.userID)
			if err != nil {
				log.Printf("Error connecting users in retry: %v", err)
				m.rdb.Del(ctx, lockKey1)
				m.rdb.Del(ctx, lockKey2)
				continue
			}

			// Remove both from searching
			m.RemoveSearchingUser(ctx, user1.userID)
			m.RemoveSearchingUser(ctx, user2.userID)
			m.rdb.Del(ctx, lockKey1)
			m.rdb.Del(ctx, lockKey2)

			// Calculate distance if both have location
			var distance float64
			var searchMode = constants.SearchModeRandom
			if user1.searchReq.Latitude != 0 && user1.searchReq.Longitude != 0 &&
				user2.searchReq.Latitude != 0 && user2.searchReq.Longitude != 0 {
				distance = calculateDistance(
					user1.searchReq.Latitude, user1.searchReq.Longitude,
					user2.searchReq.Latitude, user2.searchReq.Longitude,
				)
				searchMode = constants.SearchModeNearby
			}

			// Notify both users
			m.notifyMatch(user1.userID, user2.userID, searchMode, distance)

			matched[user1.userID] = true
			matched[user2.userID] = true

			log.Printf("✅ Retry Match: User %d <-> User %d", user1.userID, user2.userID)
			break
		}
	}

	if len(matched) > 0 {
		log.Printf("🎉 Matched %d users in retry cycle", len(matched))
	}
}

// notifyMatch mengirim notifikasi ke kedua user saat match ditemukan
func (m *Matcher) notifyMatch(userID1, userID2 int64, searchMode string, distance float64) {
	var msg1, msg2 string

	if searchMode == constants.SearchModeNearby && distance > 0 {
		distStr := formatDistance(distance)
		msg1 = fmt.Sprintf("🎉 Partner ditemukan! (📍 Jarak: *%s*)\n\nSilakan mulai percakapan.\nKetik /next untuk skip atau /stop untuk mengakhiri.", distStr)
		msg2 = msg1
	} else {
		msg1 = constants.MsgPartnerFound
		msg2 = constants.MsgPartnerFound
	}

	// Send to both users
	m.sendMessage(userID1, msg1)
	m.sendMessage(userID2, msg2)
}

// sendMessage mengirim pesan ke user
func (m *Matcher) sendMessage(userID int64, text string) {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ParseMode = "Markdown"
	_, err := m.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending match notification to %d: %v", userID, err)
	}
}

// calculateDistance menghitung jarak antara dua koordinat dalam km
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // km

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

// formatDistance format jarak dalam km
func formatDistance(distance float64) string {
	if distance < 1 {
		return "< 1 km"
	}
	return fmt.Sprintf("%.1f km", distance)
}
