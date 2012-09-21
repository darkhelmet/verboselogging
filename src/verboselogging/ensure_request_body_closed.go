package main

import "net/http"

type EnsureRequestBodyClosedHandler struct {
    h http.Handler
}

func (erbch EnsureRequestBodyClosedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()
    erbch.h.ServeHTTP(w, r)
}
