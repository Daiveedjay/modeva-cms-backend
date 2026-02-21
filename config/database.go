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
	CmsDB       *pgxpool.Pool
	EcommerceDB *pgxpool.Pool

	CmsGorm       *gorm.DB
	EcommerceGorm *gorm.DB
)

func InitDB() {
	initPgx()
	initGORM()
}

func initPgx() {
	// CMS - use Neon URL if provided
	cmsURL := os.Getenv("CMS_DB_URL")
	if cmsURL == "" {
		// fallback to local
		cmsURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/modeva_cms_backend?sslmode=disable",
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", ""),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
		)
		log.Println("⚠️ CMS_DB_URL not set, using local default")
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

	// Ecommerce - same pattern
	ecommerceURL := os.Getenv("ECOMMERCE_DB_URL")
	if ecommerceURL == "" {
		ecommerceURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/modeva_ecommerce?sslmode=disable",
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", ""),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
		)
		log.Println("⚠️ ECOMMERCE_DB_URL not set, using local default")
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

func initGORM() {
	// Shared logger config
	gormLogger := logger.Default.LogMode(logger.Info)
	if os.Getenv("APP_ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// CMS GORM: prefer full URL
	var cmsDSN string
	if os.Getenv("CMS_DB_URL") != "" {
		cmsDSN = os.Getenv("CMS_DB_URL")
	} else {
		cmsDSN = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=modeva_cms_backend port=%s sslmode=disable TimeZone=UTC",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", ""),
			getEnv("DB_PORT", "5432"),
		)
		log.Println("⚠️ CMS_DB_URL not set, using local GORM default")
	}

	var err error
	CmsGorm, err = gorm.Open(postgres.Open(cmsDSN), &gorm.Config{
		Logger:  gormLogger,
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to CMS database with GORM: %v", err)
	}
	if sqlDB, err := CmsGorm.DB(); err == nil {
		sqlDB.SetMaxOpenConns(5)
		sqlDB.SetMaxIdleConns(2)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		sqlDB.SetConnMaxIdleTime(2 * time.Minute)
	}
	log.Println("✅ CMS database connected (GORM)")

	// Ecommerce GORM: same
	var ecommerceDSN string
	if os.Getenv("ECOMMERCE_DB_URL") != "" {
		ecommerceDSN = os.Getenv("ECOMMERCE_DB_URL")
	} else {
		ecommerceDSN = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=modeva_ecommerce port=%s sslmode=disable TimeZone=UTC",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", ""),
			getEnv("DB_PORT", "5432"),
		)
		log.Println("⚠️ ECOMMERCE_DB_URL not set, using local GORM default")
	}
	EcommerceGorm, err = gorm.Open(postgres.Open(ecommerceDSN), &gorm.Config{
		Logger:  gormLogger,
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to Ecommerce database with GORM: %v", err)
	}
	if sqlDB, err := EcommerceGorm.DB(); err == nil {
		sqlDB.SetMaxOpenConns(5)
		sqlDB.SetMaxIdleConns(2)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		sqlDB.SetConnMaxIdleTime(2 * time.Minute)
	}
	log.Println("✅ Ecommerce database connected (GORM)")
}

func CloseDB() {
	if CmsDB != nil {
		CmsDB.Close()
		log.Println("✅ CMS database connection closed (pgx)")
	}
	if EcommerceDB != nil {
		EcommerceDB.Close()
		log.Println("✅ Ecommerce database connection closed (pgx)")
	}

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

// WithTimeout returns a context with a 10s timeout (bumped from 5s for Neon cold starts)
func WithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func WithCustomTimeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
