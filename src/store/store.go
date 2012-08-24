package store

import (
    "database/sql"
    "fmt"
    "github.com/darkhelmet/env"
    "sync"
    _ "vendor/github.com/bmizerany/pq"
)

var (
    user   = env.StringDefault("USER", "")
    url    = env.StringDefault("DATABASE_URL", fmt.Sprintf("dbname=darkblog2_development sslmode=disable"))
    NoRows = sql.ErrNoRows
    conn   *sql.DB
    refs   uint
    mu     sync.Mutex
)

func Connect() (*sql.DB, error) {
    mu.Lock()
    defer mu.Unlock()
    if conn == nil {
        var err error
        conn, err = sql.Open("postgres", url)
        if err != nil {
            return nil, err
        }
    }
    refs += 1
    return conn, nil
}

func Disconnect() {
    mu.Lock()
    defer mu.Unlock()
    refs -= 1
    if refs == 0 {
        conn.Close()
        conn = nil
    }
}
