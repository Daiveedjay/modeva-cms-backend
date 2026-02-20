// ════════════════════════════════════════════════════════════
// Path: config/database.go
// Database connections for CMS and Ecommerce
// ════════════════════════════════════════════════════════════

package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// pgx connections (keep for raw SQL if needed)
	CmsDB       *pgxpool.Pool
	EcommerceDB *pgxpool.Pool

	// GORM connections
	CmsGorm       *gorm.DB
	EcommerceGorm *gorm.DB
)

// InitDB initializes both database connections (pgx + GORM)
func InitDB() {
	initPgx()
	initGORM()
}

// initPgx initializes pgx connections
func initPgx() {
	// Get consistent connection parameters
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "daiveed")
	dbPort := getEnv("DB_PORT", "5432")

	// Initialize CMS database
	cmsURL := os.Getenv("CMS_DB_URL")
	if cmsURL == "" {
		cmsURL = fmt.Sprintf("postgres://%s:%s@%s:%s/modeva_cms_backend?sslmode=disable",
			dbUser, dbPassword, dbHost, dbPort)
		log.Printf("⚠️  CMS_DB_URL not set, using default")
	}

	var err error
	CmsDB, err = pgxpool.New(context.Background(), cmsURL)
	if err != nil {
		log.Fatalf("❌ Unable to connect to CMS database: %v", err)
	}

	if err = CmsDB.Ping(context.Background()); err != nil {
		log.Fatalf("❌ CMS database ping failed: %v", err)
	}

	log.Println("✅ CMS database connected (pgx)")

	// Initialize Ecommerce database
	ecommerceURL := os.Getenv("ECOMMERCE_DB_URL")
	if ecommerceURL == "" {
		ecommerceURL = fmt.Sprintf("postgres://%s:%s@%s:%s/modeva_ecommerce?sslmode=disable",
			dbUser, dbPassword, dbHost, dbPort)
		log.Printf("⚠️  ECOMMERCE_DB_URL not set, using default")
	}

	EcommerceDB, err = pgxpool.New(context.Background(), ecommerceURL)
	if err != nil {
		log.Fatalf("❌ Unable to connect to Ecommerce database: %v", err)
	}

	if err = EcommerceDB.Ping(context.Background()); err != nil {
		log.Fatalf("❌ Ecommerce database ping failed: %v", err)
	}

	log.Println("✅ Ecommerce database connected (pgx)")
}

// initGORM initializes GORM connections
func initGORM() {
	// Get environment or use defaults
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "daiveed")
	dbPort := getEnv("DB_PORT", "5432")

	// CMS Database DSN
	cmsDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, "modeva_cms_backend", dbPort,
	)

	// Configure GORM logger (show SQL in development)
	gormLogger := logger.Default.LogMode(logger.Info)
	if os.Getenv("APP_ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	var err error
	CmsGorm, err = gorm.Open(postgres.Open(cmsDSN), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to CMS database with GORM: %v", err)
	}
	log.Println("✅ CMS database connected (GORM)")

	// Ecommerce Database DSN
	ecommerceDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, "modeva_ecommerce", dbPort,
	)

	EcommerceGorm, err = gorm.Open(postgres.Open(ecommerceDSN), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to Ecommerce database with GORM: %v", err)
	}
	log.Println("✅ Ecommerce database connected (GORM)")
}

// CloseDB closes both database connections
func CloseDB() {
	if CmsDB != nil {
		CmsDB.Close()
		log.Println("✅ CMS database connection closed (pgx)")
	}
	if EcommerceDB != nil {
		EcommerceDB.Close()
		log.Println("✅ Ecommerce database connection closed (pgx)")
	}

	// Close GORM connections
	if CmsGorm != nil {
		sqlDB, _ := CmsGorm.DB()
		if sqlDB != nil {
			sqlDB.Close()
			log.Println("✅ CMS database connection closed (GORM)")
		}
	}
	if EcommerceGorm != nil {
		sqlDB, _ := EcommerceGorm.DB()
		if sqlDB != nil {
			sqlDB.Close()
			log.Println("✅ Ecommerce database connection closed (GORM)")
		}
	}
}

// WithTimeout creates a context with a 5-second timeout
func WithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// WithCustomTimeout creates a context with a custom timeout duration
func WithCustomTimeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration)
}

// getEnv gets environment variable or returns default
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
