package db

import (
	"fmt"
	"go-stock/backend/models"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Dao *gorm.DB

func Init(sqlitePath string) {
	if err := InitWithError(sqlitePath); err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}
}

// InitWithError is the startup-safe variant used by the desktop entrypoint.
// Init remains available for existing tests and auxiliary processes.
func InitWithError(sqlitePath string) error {
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second * 3,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			LogLevel:                  logger.Silent,
		},
	)
	var openDb *gorm.DB
	var err error
	if sqlitePath == "" {
		sqlitePath = "data/stock.db?_busy_timeout=10000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-524288"
	}
	openDb, err = gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{
		Logger:                                   dbLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		PrepareStmt:                              true,
	})

	if err != nil {
		return fmt.Errorf("connect sqlite: %w", err)
	}

	// 兜底：确保 busy_timeout / WAL / synchronous 生效（不同驱动/DSN 参数支持可能存在差异）
	_ = openDb.Exec("PRAGMA busy_timeout=10000").Error
	_ = openDb.Exec("PRAGMA journal_mode=WAL").Error
	_ = openDb.Exec("PRAGMA synchronous=NORMAL").Error

	dbCon, err := openDb.DB()
	if err != nil {
		return fmt.Errorf("open sqlite connection pool: %w", err)
	}
	// SQLite 写入是串行锁模型：连接开太多会放大锁竞争导致 SQLITE_BUSY
	dbCon.SetMaxIdleConns(1)
	dbCon.SetMaxOpenConns(5)
	dbCon.SetConnMaxLifetime(time.Hour)
	Dao = openDb
	if err := openDb.AutoMigrate(&ChatMemory{}, &models.StockChangeHistory{}, &models.MarketStatistic{}); err != nil {
		return fmt.Errorf("migrate core database tables: %w", err)
	}
	return nil
}

// BackupSQLite creates a transactionally consistent SQLite backup before a
// schema upgrade. VACUUM INTO also includes committed WAL pages.
func BackupSQLite(database *gorm.DB, sqlitePath, label string) (string, error) {
	if database == nil {
		return "", fmt.Errorf("database is nil")
	}
	cleanPath := strings.TrimSpace(sqlitePath)
	if query := strings.IndexByte(cleanPath, '?'); query >= 0 {
		cleanPath = cleanPath[:query]
	}
	if cleanPath == "" || cleanPath == ":memory:" {
		return "", nil
	}
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("inspect sqlite database: %w", err)
	}
	if info.Size() == 0 {
		return "", nil
	}
	backupDir := filepath.Join(filepath.Dir(cleanPath), "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}
	if strings.TrimSpace(label) == "" {
		label = "schema-upgrade"
	}
	label = strings.NewReplacer("/", "-", "\\", "-", ":", "-").Replace(label)
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s-%s-%s.db", strings.TrimSuffix(filepath.Base(cleanPath), filepath.Ext(cleanPath)), label, time.Now().Format("20060102-150405.000")))
	escapedPath := strings.ReplaceAll(filepath.ToSlash(backupPath), "'", "''")
	if err := database.Exec(fmt.Sprintf("VACUUM INTO '%s'", escapedPath)).Error; err != nil {
		return "", fmt.Errorf("backup sqlite database: %w", err)
	}
	return backupPath, nil
}
