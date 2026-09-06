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

// Connection pool sizing. The CMS pool is larger because the admin dashboard
// runs several concurrent queries per page; the storefront pool is smaller and
// shorter-lived because its traffic is bursty and mostly cached upstream.
const (
	cmsMaxOpenConns    = 10
	cmsMaxIdleConns    = 5
	cmsConnMaxLifetime = 30 * time.Minute
	cmsConnMaxIdleTime = 10 * time.Minute

	ecommerceMaxOpenConns    = 5
	ecommerceMaxIdleConns    = 2
	ecommerceConnMaxLifetime = 5 * time.Minute
	ecommerceConnMaxIdleTime = 2 * time.Minute

	defaultQueryTimeout = 5 * time.Second
)

// Local fallbacks, used only when no *_DB_URL is provided.
const (
	defaultDBHost = "localhost"
	defaultDBPort = "5432"
	defaultDBUser = "postgres"
)

var (
	CmsDB       *pgxpool.Pool
	EcommerceDB *pgxpool.Pool

	CmsGorm       *gorm.DB
	EcommerceGorm *gorm.DB
)

// poolSettings groups the four numbers a GORM pool needs, so the CMS and
// storefront databases are configured by the same code path rather than by two
// blocks that have to be kept in step by hand.
type poolSettings struct {
	maxOpen     int
	maxIdle     int
	maxLifetime time.Duration
	maxIdleTime time.Duration
}

func InitDB() {
	initPgx()
	initGORM()
}

// pgxURL returns the configured URL for a database, or a local fallback.
func pgxURL(urlEnv, dbName string) string {
	if url := os.Getenv(urlEnv); url != "" {
		return url
	}
	log.Printf("⚠️ %s not set, using local default", urlEnv)
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("DB_USER", defaultDBUser),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_HOST", defaultDBHost),
		getEnv("DB_PORT", defaultDBPort),
		dbName,
	)
}

// gormDSN returns the configured URL for a database, or a local key=value DSN.
func gormDSN(urlEnv, dbName string) string {
	if url := os.Getenv(urlEnv); url != "" {
		return url
	}
	log.Printf("⚠️ %s not set, using local GORM default", urlEnv)
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("DB_HOST", defaultDBHost),
		getEnv("DB_USER", defaultDBUser),
		getEnv("DB_PASSWORD", ""),
		dbName,
		getEnv("DB_PORT", defaultDBPort),
	)
}

func connectPgx(label, urlEnv, dbName string) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), pgxURL(urlEnv, dbName))
	if err != nil {
		log.Fatalf("❌ Unable to connect to %s database: %v", label, err)
	}
	if err = pool.Ping(context.Background()); err != nil {
		log.Fatalf("❌ %s database ping failed: %v", label, err)
	}
	log.Printf("✅ %s database connected (pgx)", label)
	return pool
}

func initPgx() {
	CmsDB = connectPgx("CMS", "CMS_DB_URL", "modeva_cms_backend")
	EcommerceDB = connectPgx("Ecommerce", "ECOMMERCE_DB_URL", "modeva_ecommerce")
}

func connectGorm(label, urlEnv, dbName string, gormLogger logger.Interface, settings poolSettings) *gorm.DB {
	db, err := gorm.Open(postgres.Open(gormDSN(urlEnv, dbName)), &gorm.Config{
		Logger:  gormLogger,
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to %s database with GORM: %v", label, err)
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(settings.maxOpen)
		sqlDB.SetMaxIdleConns(settings.maxIdle)
		sqlDB.SetConnMaxLifetime(settings.maxLifetime)
		sqlDB.SetConnMaxIdleTime(settings.maxIdleTime)
	}
	log.Printf("✅ %s database connected (GORM)", label)
	return db
}

func initGORM() {
	gormLogger := logger.Default.LogMode(logger.Info)
	if os.Getenv("APP_ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	CmsGorm = connectGorm("CMS", "CMS_DB_URL", "modeva_cms_backend", gormLogger, poolSettings{
		maxOpen:     cmsMaxOpenConns,
		maxIdle:     cmsMaxIdleConns,
		maxLifetime: cmsConnMaxLifetime,
		maxIdleTime: cmsConnMaxIdleTime,
	})
	EcommerceGorm = connectGorm("Ecommerce", "ECOMMERCE_DB_URL", "modeva_ecommerce", gormLogger, poolSettings{
		maxOpen:     ecommerceMaxOpenConns,
		maxIdle:     ecommerceMaxIdleConns,
		maxLifetime: ecommerceConnMaxLifetime,
		maxIdleTime: ecommerceConnMaxIdleTime,
	})
}

func closePgx(pool *pgxpool.Pool, label string) {
	if pool == nil {
		return
	}
	pool.Close()
	log.Printf("✅ %s database connection closed (pgx)", label)
}

func closeGorm(db *gorm.DB, label string) {
	if db == nil {
		return
	}
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
		log.Printf("✅ %s database connection closed (GORM)", label)
	}
}

func CloseDB() {
	closePgx(CmsDB, "CMS")
	closePgx(EcommerceDB, "Ecommerce")
	closeGorm(CmsGorm, "CMS")
	closeGorm(EcommerceGorm, "Ecommerce")
}

// WithTimeout returns a context using the default query timeout.
func WithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultQueryTimeout)
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
