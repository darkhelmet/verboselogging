package main

import (
    "log"
    "net/http"
)

type LoggerHandler struct {
    h      http.Handler
    logger *log.Logger
}

func (lh LoggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    lh.logger.Printf("%s %s\n", r.Method, r.URL)
    lh.h.ServeHTTP(w, r)
}
