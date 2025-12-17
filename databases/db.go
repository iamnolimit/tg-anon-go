package databases

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"tg-anon-go/constants"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

// resolveIPv4 resolves hostname to IPv4 address
func resolveIPv4(host string) (string, error) {
	// Use Google DNS to resolve IPv4 only
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 10 * time.Second}
			return d.DialContext(ctx, "udp4", "8.8.8.8:53")
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ips, err := resolver.LookupIP(ctx, "ip4", host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s to IPv4: %v", host, err)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IPv4 addresses found for %s", host)
	}

	log.Printf("🔍 Resolved %s to IPv4: %s", host, ips[0].String())
	return ips[0].String(), nil
}

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
	config.MaxConns = 5
	config.MinConns = 1

	// Pre-resolve the host to IPv4
	originalHost := config.ConnConfig.Host
	ipv4Addr, err := resolveIPv4(originalHost)
	if err != nil {
		return fmt.Errorf("failed to resolve host: %v", err)
	}

	// Replace hostname with IPv4 address in config
	config.ConnConfig.Host = ipv4Addr

	// Force IPv4 for all connections
	config.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		d := net.Dialer{Timeout: 30 * time.Second}
		conn, err := d.DialContext(ctx, "tcp4", addr)
		if err != nil {
			log.Printf("❌ Dial error: %v", err)
			return nil, err
		}
		log.Printf("✅ Connected via IPv4: %s", addr)
		return conn, nil
	}

	log.Printf("🔌 Connecting to NeonDB at %s...", ipv4Addr)

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("failed to create pool: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	DB = pool
	log.Println("✅ Connected to NeonDB successfully!")

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
