package main

import (
    "net/http"
    "net/url"
)

type CanonicalHostHandler struct {
    h             http.Handler
    canonicalHost string
    scheme        string
}

func (chh CanonicalHostHandler) replaceHost(u url.URL) string {
    u.Host = chh.canonicalHost
    u.Scheme = chh.scheme
    return u.String()
}

func (chh CanonicalHostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Host != chh.canonicalHost {
        u := chh.replaceHost(*r.URL)
        logger.Printf("redirecting to %s", u)
        http.Redirect(w, r, u, http.StatusMovedPermanently)
        return
    }
    chh.h.ServeHTTP(w, r)
}
