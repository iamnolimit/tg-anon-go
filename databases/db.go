package databases

import (
	"context"
	"log"
	"os"

	"tg-anon-go/constants"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

// InitDatabase menginisialisasi koneksi ke NeonDB
func InitDatabase() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return err
	}

	// Set pool configuration
	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return err
	}

	DB = pool
	log.Println("✅ Connected to NeonDB successfully")

	// Run migrations
	if err := runMigrations(); err != nil {
		return err
	}

	return nil
}

// runMigrations menjalankan migrasi database
func runMigrations() error {
	ctx := context.Background()

	// Create users table
	if _, err := DB.Exec(ctx, constants.QueryCreateUsersTable); err != nil {
		log.Printf("Error creating users table: %v", err)
		return err
	}

	// Create sessions table
	if _, err := DB.Exec(ctx, constants.QueryCreateSessionsTable); err != nil {
		log.Printf("Error creating sessions table: %v", err)
		return err
	}
	// Create messages table
	if _, err := DB.Exec(ctx, constants.QueryCreateMessagesTable); err != nil {
		log.Printf("Error creating messages table: %v", err)
		return err
	}

	// Create vars table
	if _, err := DB.Exec(ctx, constants.QueryCreateVarsTable); err != nil {
		log.Printf("Error creating vars table: %v", err)
		return err
	}

	log.Println("✅ Database migrations completed")
	return nil
}

// CloseDatabase menutup koneksi database
func CloseDatabase() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}
