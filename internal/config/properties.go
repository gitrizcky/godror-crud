package config

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"
    "sync/atomic"
    "time"
    "log"
)

// Queries holds all SQL statements used by the repo.
type Queries struct {
    ListProducts  string
    GetProduct    string
    NextProductID string
    InsertProduct string
    UpdateProduct string
    DeleteProduct string
}

// defaultQueries provides built-in fallbacks if properties are missing.
var defaultQueries = Queries{
    ListProducts:  "SELECT PRODUCT_ID, NAME, ATTRIBUTES_VARCHAR, ATTRIBUTES_CLOB, ATTRIBUTES_BLOB FROM PRODUCTS ORDER BY PRODUCT_ID",
    GetProduct:    "SELECT NAME, ATTRIBUTES_VARCHAR, ATTRIBUTES_CLOB, ATTRIBUTES_BLOB FROM PRODUCTS WHERE PRODUCT_ID = :1",
    NextProductID: "SELECT NVL(MAX(PRODUCT_ID), 0) + 1 FROM PRODUCTS",
    InsertProduct: "INSERT INTO PRODUCTS (PRODUCT_ID, NAME, ATTRIBUTES_VARCHAR, ATTRIBUTES_CLOB, ATTRIBUTES_BLOB) VALUES (:1, :2, :3, :4, :5)",
    UpdateProduct: "UPDATE PRODUCTS SET NAME = :1, ATTRIBUTES_VARCHAR = :2, ATTRIBUTES_CLOB = :3, ATTRIBUTES_BLOB = :4 WHERE PRODUCT_ID = :5",
    DeleteProduct: "DELETE FROM PRODUCTS WHERE PRODUCT_ID = :1",
}

// Manager watches a .properties file and exposes the latest Queries.
type Manager struct {
    path     string
    current  atomic.Value // stores Queries
}

// NewManager creates a Manager with defaults loaded.
func NewManager(path string) *Manager {
    m := &Manager{path: path}
    m.current.Store(defaultQueries)
    return m
}

// Get returns the currently active queries.
func (m *Manager) Get() Queries {
    return m.current.Load().(Queries)
}

// Start begins a polling loop that reloads the properties file when it changes.
// It polls the mtime every 2 seconds. Call in a goroutine.
func (m *Manager) Start(stop <-chan struct{}) {
    var lastModTime time.Time

    // Initial load (if file exists)
    if fi, err := os.Stat(m.path); err == nil {
        if q, err := loadQueries(m.path); err == nil {
            m.current.Store(q)
            lastModTime = fi.ModTime()
        } else {
            log.Printf("config: initial load error, using defaults: %v", err)
        }
    }

    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-stop:
            return
        case <-ticker.C:
            fi, err := os.Stat(m.path)
            if err != nil {
                // If file disappears, keep last known good config.
                continue
            }
            if fi.ModTime().After(lastModTime) {
                if q, err := loadQueries(m.path); err == nil {
                    m.current.Store(q)
                    lastModTime = fi.ModTime()
                    log.Printf("config: reloaded queries from %s", filepath.Base(m.path))
                } else {
                    log.Printf("config: reload failed (keeping previous): %v", err)
                }
            }
        }
    }
}

// loadQueries parses a Java-style properties file and maps well-known keys
// to the Queries struct. Unknown/missing keys fall back to defaults.
func loadQueries(path string) (Queries, error) {
    f, err := os.Open(path)
    if err != nil {
        return Queries{}, err
    }
    defer f.Close()

    values := map[string]string{}
    s := bufio.NewScanner(f)
    for s.Scan() {
        line := strings.TrimSpace(s.Text())
        if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
            continue
        }
        // key=value default; if '=' is absent, allow 'key: value'
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
        values[strings.ToLower(key)] = val
    }
    if err := s.Err(); err != nil {
        return Queries{}, err
    }

    q := defaultQueries
    if v, ok := values["listproducts"]; ok {
        q.ListProducts = v
    }
    if v, ok := values["getproduct"]; ok {
        q.GetProduct = v
    }
    if v, ok := values["nextproductid"]; ok {
        q.NextProductID = v
    }
    if v, ok := values["insertproduct"]; ok {
        q.InsertProduct = v
    }
    if v, ok := values["updateproduct"]; ok {
        q.UpdateProduct = v
    }
    if v, ok := values["deleteproduct"]; ok {
        q.DeleteProduct = v
    }
    return q, nil
}
