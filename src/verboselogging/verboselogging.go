package main

import (
    "bytes"
    "compress/gzip"
    "fmt"
    "github.com/darkhelmet/env"
    "io"
    "log"
    "net"
    "os"
    Page "page"
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

func withGzip(req *web.Request, status int, contentType string, f func(io.Writer)) {
    if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
        w := req.Respond(status, web.HeaderContentType, contentType,
            web.HeaderContentEncoding, "gzip",
            web.HeaderVary, "Accept-Encoding")
        gz := gzip.NewWriter(w)
        defer gz.Close()
        f(gz)
    } else {
        w := req.Respond(status, web.HeaderContentType, contentType)
        f(w)
    }
}

func rootHandler(req *web.Request) {
    withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
        view.RenderLayout(w, "rootHandler", "/", "Verbose Logging", "software development with some really amazing hair")
    })
}

func opensearchHandler(req *web.Request) {
    withGzip(req, web.StatusOK, "application/xml; charset=utf-8", func(w io.Writer) {
        view.RenderPartial(w, "opensearch.tmpl", nil)
    })
}

func searchHandler(req *web.Request) {
    withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
        io.WriteString(w, "searchHandler")
    })
}

func feedHandler(req *web.Request) {
    withGzip(req, web.StatusOK, "application/rss+xml; charset=utf-8", func(w io.Writer) {
        io.WriteString(w, "feedHandler")
    })
}

func sitemapHandler(req *web.Request) {
    withGzip(req, web.StatusOK, "application/xml; charset=utf-8", func(w io.Writer) {
        view.RenderPartial(w, "sitemapHandler", nil)
    })
}

func archiveHandler(req *web.Request) {
    io.WriteString(w, "archiveHandler")
}

func monthlyHandler(req *web.Request) {
    io.WriteString(w, "monthlyHandler")
}

func categoryHandler(req *web.Request) {
    io.WriteString(w, "categoryHandler")
}

func permalinkHandler(req *web.Request) {
    io.WriteString(w, "permalinkHandler")
}

func tagHandler(req *web.Request) {
    io.WriteString(w, "tagHandler")
}

func pageHandler(req *web.Request) {
    page := Page.FindBySlug(req.URLParam["slug"])
    if page == nil {
        notFound(w)
    } else {
        io.WriteString(w, "pageHandler")
    }
}

func notFound(w io.Writer) {
    var buffer bytes.Buffer
    view.RenderPartial(&buffer, "not_found.tmpl", nil)
    view.RenderLayout(w, buffer.String(), "", "", "")
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
        Register("/", "GET", rootHandler).
        Register("/opensearch.xml", "GET", opensearchHandler).
        Register("/search", "GET", searchHandler).
        Register("/feed", "GET", feedHandler).
        Register("/sitemap.xml<gzip:(\\.gz)?>", "GET", withGzip("application/xml; charset=utf-8", sitemapHandler)).
        Register("/archive/<archive:(full|category|month)>", "GET", withGzip(archiveHandler)).
        Register("/<year:\\d{4}>/<month:\\d{2}>", "GET", monthlyHandler).
        Register("/category/<category>", "GET", categoryHandler).
        Register("/<year:\\d{4}>/<month:\\d{2}>/<day:\\d{2}>/<slug>", "GET", permalinkHandler).
        Register("/tag/<tag>", "GET", tagHandler).
        Register("/<slug:\\w+>", "GET", pageHandler).
        Register("/<path:.*>", "GET", web.DirectoryHandler("public", nil))

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
