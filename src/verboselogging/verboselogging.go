package main

import (
    "compress/gzip"
    "fmt"
    "github.com/darkhelmet/env"
    "io"
    "log"
    "net"
    "os"
    "strings"
    "vendor/github.com/garyburd/twister/server"
    "vendor/github.com/garyburd/twister/web"
    "view"
)

var (
    port          = env.IntDefault("PORT", 8080)
    canonicalHost = env.StringDefaultF("CANONICAL_HOST", func() string { return fmt.Sprintf("localhost:%d", port) })
    logger        = log.New(os.Stdout, "[server] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
)

type h func(*web.Request, io.Writer)

func withGzip(contentType string, f h) func(*web.Request) {
    return func(req *web.Request) {
        if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
            w := req.Respond(web.StatusOK, web.HeaderContentType, contentType, web.HeaderContentEncoding, "gzip")
            gz := gzip.NewWriter(w)
            defer gz.Close()
            f(req, gz)
        } else {
            w := req.Respond(web.StatusOK, web.HeaderContentType, contentType)
            f(req, w)
        }
    }
}

func rootHandler(req *web.Request, w io.Writer) {
    view.RenderLayout(w, "rootHandler", "Verbose Logging", "software development with some really amazing hair")
}

func opensearchHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "opensearchHandler")
}

func searchHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "searchHandler")
}

func feedHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "feedHandler")
}

func sitemapHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "sitemapHandler")
}

func archiveHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "archiveHandler")
}

func monthlyHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "monthlyHandler")
}

func categoryHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "categoryHandler")
}

func permalinkHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "permalinkHandler")
}

func tagHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "tagHandler")
}

func pageHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "pageHandler")
}

func notFoundHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "notFoundHandler")
}

func redirectHandler(req *web.Request) {
    url := req.URL
    url.Host = canonicalHost
    url.Scheme = "http"
    req.Respond(web.StatusMovedPermanently, web.HeaderLocation, url.String())
}

func ShortLogger(lr *server.LogRecord) {
    if lr.Error != nil {
        logger.Printf("%d %s %s %s\n", lr.Status, lr.Request.Method, lr.Request.URL, lr.Error)
    } else {
        logger.Printf("%d %s %s\n", lr.Status, lr.Request.Method, lr.Request.URL)
    }
}

func main() {
    router := web.NewRouter().
        Register("/", "GET", withGzip("text/html; charset=utf-8", rootHandler)).
        Register("/opensearch.xml", "GET", opensearchHandler).
        Register("/search", "GET", searchHandler).
        Register("/feed", "GET", feedHandler).
        Register("/sitemap.xml<gzip:(\\.gz)?>", "GET", sitemapHandler).
        Register("/archive/<archive:(full|category|month)>", "GET", archiveHandler).
        Register("/<year:\\d{4}>/<month:\\d{2}>", "GET", monthlyHandler).
        Register("/category/<category>", "GET", categoryHandler).
        Register("/<year:\\d{4}>/<month:\\d{2}>/<day:\\d{2}>/<slug>", "GET", permalinkHandler).
        Register("/tag/<tag>", "GET", tagHandler).
        Register("/<page>", "GET", pageHandler).
        Register("/<splat:.*>", "GET", notFoundHandler)

    redirector := web.NewRouter().
        Register("/<splat:.*>", "GET", redirectHandler)

    hostRouter := web.NewHostRouter(redirector).
        Register(canonicalHost, router)

    listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
    if err != nil {
        logger.Fatalf("Failed to listen: %s", err)
    }
    defer listener.Close()
    server := &server.Server{
        Listener: listener,
        Handler:  hostRouter,
        Logger:   server.LoggerFunc(ShortLogger),
    }
    logger.Printf("verboselogging is starting on 0.0.0.0:%d", port)
    err = server.Serve()
    if err != nil {
        logger.Fatalf("Failed to serve: %s", err)
    }
}
