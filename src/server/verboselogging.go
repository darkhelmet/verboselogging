package main

import (
    "config"
    "fmt"
    "log"
    "net/http"
    "os"
    "verboselogging"
)

var (
    logger = log.New(os.Stdout, "[server] ", config.LogFlags)
)

func main() {
    handler := verboselogging.SetupHandler()
    http.Handle("/", handler)
    logger.Printf("verboselogging is starting on 0.0.0.0:%d", config.Port)
    err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", config.Port), nil)
    if err != nil {
        logger.Fatalf("Failed to serve: %s", err)
    }
}
