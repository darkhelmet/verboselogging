package store

import (
    "config"
    "database/sql"
    "sync"
    _ "vendor/github.com/bmizerany/pq"
    "vendor/github.com/bmizerany/pq"
)

var (
    NoRows = sql.ErrNoRows
    conn   *sql.DB
    refs   uint
    mu     sync.Mutex
    url    string
)

func init() {
    var err error
    url, err = pq.ParseURL(config.DatabaseUrl)
    if err != nil {
        panic(err)
    }
}

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
