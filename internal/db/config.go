package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Config holds database connection parameters.
type Config struct {
	Service  string
	Username string
	Server   string
	Port     string
	Password string
	Timezone string // e.g. "+00:00" or IANA name

	// Optional pool settings (database/sql)
	MaxOpenConns    int           // 0 = unlimited
	MaxIdleConns    int           // 0 = default
	ConnMaxLifetime time.Duration // 0 = forever
	ConnMaxIdleTime time.Duration // 0 = forever

	// Optional: godror driver session pooling
	// If StandaloneConnection is explicitly set to false (0), godror uses the session pool.
	// When nil, the parameter is omitted and driver default applies.
	StandaloneConnection *bool
	PoolMinSessions      int // 0 = omit
	PoolMaxSessions      int // 0 = omit
	PoolIncrement        int // 0 = omit
}

// Open creates a connection pool to Oracle using godror.
func Open(cfg Config) (*sql.DB, error) {
	tz := cfg.Timezone
	if tz == "" {
		tz = "+00:00"
	}
	dsn := fmt.Sprintf(
		"user=\"%s\" password=\"%s\" connectString=\"%s:%s/%s\" timezone=\"%s\"",
		cfg.Username, cfg.Password, cfg.Server, cfg.Port, cfg.Service, tz,
	)
	// Append optional driver session pooling params
	if cfg.StandaloneConnection != nil {
		if *cfg.StandaloneConnection {
			dsn += " standaloneConnection=1"
		} else {
			dsn += " standaloneConnection=0"
		}
	}
	if cfg.PoolMinSessions > 0 {
		dsn += fmt.Sprintf(" poolMinSessions=%d", cfg.PoolMinSessions)
	}
	if cfg.PoolMaxSessions > 0 {
		dsn += fmt.Sprintf(" poolMaxSessions=%d", cfg.PoolMaxSessions)
	}
	if cfg.PoolIncrement > 0 {
		dsn += fmt.Sprintf(" poolIncrement=%d", cfg.PoolIncrement)
	}
	d, err := sql.Open("godror", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// Apply pool settings if provided
	if cfg.MaxOpenConns > 0 {
		d.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		d.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		d.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		d.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}
	if err := d.Ping(); err != nil {
		_ = d.Close()
		return nil, fmt.Errorf("db.Ping: %w", err)
	}
	return d, nil
}
