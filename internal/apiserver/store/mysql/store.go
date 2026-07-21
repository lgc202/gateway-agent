// Package mysql 封装 apiserver 对 MySQL 的数据访问
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"

	"github.com/lgc202/gateway-agent/internal/apiserver/config"
	"github.com/lgc202/gateway-agent/internal/apiserver/store/mysql/sqlc"
)

const (
	databaseConnectTimeout   = 5 * time.Second
	databaseReadWriteTimeout = 10 * time.Second
	databasePingTimeout      = 5 * time.Second
	databaseOperationTimeout = 10 * time.Second
)

// Store 提供 Chat 领域当前所需的数据库操作
type Store struct {
	db      *sql.DB
	queries *sqlc.Queries
}

// NewDB 创建数据库连接池，并在启动阶段验证 MySQL 可用性
func NewDB(cfg *config.Config) (*sql.DB, error) {
	driverConfig := mysqldriver.NewConfig()
	driverConfig.User = cfg.MySQL.Username
	driverConfig.Passwd = cfg.MySQL.Password
	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(cfg.MySQL.Host, strconv.Itoa(cfg.MySQL.Port))
	driverConfig.DBName = cfg.MySQL.Database
	driverConfig.ParseTime = true
	driverConfig.Loc = time.UTC
	driverConfig.Collation = "utf8mb4_0900_ai_ci"
	driverConfig.Params = map[string]string{"charset": "utf8mb4"}
	driverConfig.Timeout = databaseConnectTimeout
	driverConfig.ReadTimeout = databaseReadWriteTimeout
	driverConfig.WriteTimeout = databaseReadWriteTimeout

	db, err := sql.Open("mysql", driverConfig.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open MySQL: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), databasePingTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping MySQL: %w", err)
	}

	return db, nil
}

// NewStore 创建 MySQL Store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		queries: sqlc.New(db),
	}
}

// WithinTransaction 在 Read Committed 隔离级别中执行一组数据库操作
func (s *Store) WithinTransaction(ctx context.Context, fn func(context.Context, *sqlc.Queries) error) error {
	ctx, cancel := context.WithTimeout(ctx, databaseOperationTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return databaseError(err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := fn(ctx, s.queries.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return databaseError(err)
	}

	return nil
}
