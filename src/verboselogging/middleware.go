package main

import (
    "compress/gzip"
    "io"
    "strings"
    "vendor/github.com/garyburd/twister/web"
)

type Responder interface {
    Respond(status int, headerKeysAndValues ...string) io.Writer
}

type Middleware func(*web.Request, Responder)

func Use(f Middleware) func(*web.Request) {
    return func(req *web.Request) {
        f(req, req)
    }
}

type gzipResponder struct {
    up  Responder
    gz  *gzip.Writer
}

func (r *gzipResponder) Close() {
    r.gz.Close()
}

func (r *gzipResponder) Respond(status int, headerKeysAndValues ...string) io.Writer {
    headerKeysAndValues = append(headerKeysAndValues,
        web.HeaderContentEncoding, "gzip",
        web.HeaderVary, "Accept-Encoding")
    w := r.up.Respond(status, headerKeysAndValues...)
    r.gz = gzip.NewWriter(w)
    return r.gz
}

func Gzip(f Middleware) Middleware {
    return func(req *web.Request, r Responder) {
        if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
            gzr := &gzipResponder{up: r}
            defer gzr.Close()
            f(req, gzr)
        } else {
            f(req, r)
        }
    }
}
