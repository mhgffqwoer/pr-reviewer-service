package postgres

import (
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mhgffqwoer/pr-reviewer-service/internal/config"
)

var (
	pool *sqlx.DB
	once sync.Once
)

func Connect(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	var err error
	once.Do(func() {
		localPool, openErr := sqlx.Open("postgres", cfg.URL)
		if openErr != nil {
			err = fmt.Errorf("failed to connect to postgres: %w", openErr)
			return
		}

		localPool.SetMaxOpenConns(cfg.MaxConnections)
		localPool.SetMaxIdleConns(cfg.MaxIdleConnections)
		localPool.SetConnMaxLifetime(time.Duration(cfg.ConnectionMaxLifetime) * time.Minute)

		if pingErr := localPool.Ping(); pingErr != nil {
			err = fmt.Errorf("DB ping failed: %w", pingErr)
			return
		}
		pool = localPool
	})
	return pool, err
}
