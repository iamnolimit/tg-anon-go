package constants

// Database Tables
const (
	TableUsers    = "users"
	TableSessions = "chat_sessions"
	TableMessages = "messages"
	TableVars     = "vars"
)

// SQL Queries
const (
	QueryCreateUsersTable = `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			telegram_id BIGINT UNIQUE NOT NULL,
			username VARCHAR(255),
			first_name VARCHAR(255),
			status VARCHAR(50) DEFAULT 'idle',
			partner_id BIGINT DEFAULT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
		CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
	`

	QueryCreateSessionsTable = `
		CREATE TABLE IF NOT EXISTS chat_sessions (
			id SERIAL PRIMARY KEY,
			user1_id BIGINT NOT NULL,
			user2_id BIGINT NOT NULL,
			started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ended_at TIMESTAMP DEFAULT NULL,
			is_active BOOLEAN DEFAULT TRUE
		);
		CREATE INDEX IF NOT EXISTS idx_sessions_active ON chat_sessions(is_active);
	`
	QueryCreateMessagesTable = `
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			session_id INT REFERENCES chat_sessions(id),
			sender_id BIGINT NOT NULL,
			receiver_id BIGINT NOT NULL,
			message_type VARCHAR(50) DEFAULT 'text',
			content TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	QueryCreateVarsTable = `
		CREATE TABLE IF NOT EXISTS vars (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			var_key VARCHAR(255) NOT NULL,
			var_value TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, var_key)
		);
		CREATE INDEX IF NOT EXISTS idx_vars_user_key ON vars(user_id, var_key);
	`
)
