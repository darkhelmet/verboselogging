package main

import (
    "compress/gzip"
    "net/http"
    "strings"
)

const (
    HeaderAcceptEncoding  = "Accept-Encoding"
    HeaderContentEncoding = "Content-Encoding"
    HeaderVary            = "Vary"
)

type GzipResponseWriter struct {
    gzr *gzip.Writer
    w   http.ResponseWriter
}

func (grw GzipResponseWriter) Header() http.Header {
    return grw.w.Header()
}

func (grw GzipResponseWriter) WriteHeader(code int) {
    grw.w.WriteHeader(code)
}

func (grw GzipResponseWriter) Write(b []byte) (int, error) {
    return grw.gzr.Write(b)
}

type GzipHandler struct {
    h http.Handler
}

func (gh GzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if strings.Contains(r.Header.Get(HeaderAcceptEncoding), "gzip") {
        w.Header().Set(HeaderContentEncoding, "gzip")
        w.Header().Set(HeaderVary, HeaderAcceptEncoding)
        gz := gzip.NewWriter(w)
        defer gz.Close()
        gzw := GzipResponseWriter{gz, w}
        gh.h.ServeHTTP(gzw, r)
        return
    }

    gh.h.ServeHTTP(w, r)
}
