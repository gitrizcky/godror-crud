package config

import (
    "bufio"
    "errors"
    "os"
    "strconv"
    "strings"
    "time"

    "go-demo-crud/internal/db"
)

// LoadDBConfig reads database settings from a Java-style properties file.
// Expected keys (case-insensitive):
//   db.service, db.username, db.password, db.server, db.port, db.timezone
func LoadDBConfig(path string) (db.Config, error) {
    props, err := readProperties(path)
    if err != nil {
        return db.Config{}, err
    }
    // Normalize keys to lowercase
    lc := map[string]string{}
    for k, v := range props {
        lc[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
    }

    cfg := db.Config{
        Service:  lc["db.service"],
        Username: lc["db.username"],
        Password: lc["db.password"],
        Server:   lc["db.server"],
        Port:     lc["db.port"],
        Timezone: lc["db.timezone"],
    }

    // Optional pool settings
    if v := lc["db.pool.max_open"]; v != "" {
        if n, err := strconv.Atoi(v); err == nil { cfg.MaxOpenConns = n }
    }
    if v := lc["db.pool.max_idle"]; v != "" {
        if n, err := strconv.Atoi(v); err == nil { cfg.MaxIdleConns = n }
    }
    if v := lc["db.pool.conn_max_lifetime"]; v != "" {
        if d, err := time.ParseDuration(v); err == nil { cfg.ConnMaxLifetime = d }
    }
    if v := lc["db.pool.conn_max_idletime"]; v != "" {
        if d, err := time.ParseDuration(v); err == nil { cfg.ConnMaxIdleTime = d }
    }

    // Optional: godror session pooling
    if v := lc["db.godror.standalone_connection"]; v != "" {
        if b, ok := parseBool01(v); ok {
            cfg.StandaloneConnection = &b
        }
    }
    if v := lc["db.godror.pool_min_sessions"]; v != "" {
        if n, err := strconv.Atoi(v); err == nil { cfg.PoolMinSessions = n }
    }
    if v := lc["db.godror.pool_max_sessions"]; v != "" {
        if n, err := strconv.Atoi(v); err == nil { cfg.PoolMaxSessions = n }
    }
    if v := lc["db.godror.pool_increment"]; v != "" {
        if n, err := strconv.Atoi(v); err == nil { cfg.PoolIncrement = n }
    }

    // Minimal validation
    if cfg.Service == "" || cfg.Username == "" || cfg.Password == "" || cfg.Server == "" || cfg.Port == "" {
        return cfg, errors.New("incomplete DB config in properties file")
    }
    if cfg.Timezone == "" {
        cfg.Timezone = "+00:00"
    }
    return cfg, nil
}

// readProperties loads key=value lines from a file, ignoring comments and blanks.
func readProperties(path string) (map[string]string, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    out := map[string]string{}
    s := bufio.NewScanner(f)
    for s.Scan() {
        line := strings.TrimSpace(s.Text())
        if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
            continue
        }
        sep := "="
        if !strings.Contains(line, "=") && strings.Contains(line, ":") {
            sep = ":"
        }
        parts := strings.SplitN(line, sep, 2)
        if len(parts) != 2 {
            continue
        }
        key := strings.TrimSpace(parts[0])
        val := strings.TrimSpace(parts[1])
        out[key] = val
    }
    if err := s.Err(); err != nil {
        return nil, err
    }
    return out, nil
}

// parseBool01 returns a bool for common boolean encodings and whether parsing succeeded.
func parseBool01(s string) (bool, bool) {
    switch strings.ToLower(strings.TrimSpace(s)) {
    case "1", "true", "t", "yes", "y", "on":
        return true, true
    case "0", "false", "f", "no", "n", "off":
        return false, true
    default:
        return false, false
    }
}
