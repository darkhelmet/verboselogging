package main

import (
    "bytes"
    "compress/gzip"
    "config"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    Page "page"
    "runtime"
    "strings"
    "vendor/github.com/garyburd/twister/server"
    "vendor/github.com/garyburd/twister/web"
    "view"
)

var (
    logger = log.New(os.Stdout, "[server] ", config.LogFlags)
)

type h func(*web.Request, io.Writer)

func init() {
    if config.Multicore {
        runtime.GOMAXPROCS(runtime.NumCPU())
    }
}

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
        view.RenderLayout(w, &view.RenderInfo{Yield: "rootHandler", Canonical: "/"})
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
    withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
        io.WriteString(w, "archiveHandler")
    })
}

func monthlyHandler(req *web.Request) {
    withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
        io.WriteString(w, "monthlyHandler")
    })
}

func categoryHandler(req *web.Request) {
    withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
        io.WriteString(w, "categoryHandler")
    })
}

func permalinkHandler(req *web.Request) {
    withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
        io.WriteString(w, "permalinkHandler")
    })
}

func tagHandler(req *web.Request) {
    withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
        io.WriteString(w, "tagHandler")
    })
}

func pageHandler(req *web.Request) {
    slug := req.URLParam["slug"]
    page, err := Page.FindBySlug(slug)
    if err != nil {
        switch err.(type) {
        case Page.NotFound:
            withGzip(req, web.StatusNotFound, "text/html; charset=utf-8", func(w io.Writer) {
                notFound(w)
            })
        default:
            logger.Printf("failed finding page with slug %#v: %s (%T)", slug, err, err)
            withGzip(req, web.StatusInternalServerError, "text/html; charset=utf-8", func(w io.Writer) {
                serverError(w)
            })
        }
    } else {
        withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
            view.RenderLayout(w, &view.RenderInfo{
                Yield:     view.HTML(page.BodyHtml),
                Title:     page.Title,
                Canonical: page.Canonical(),
            })
        })
    }
}

func serverError(w io.Writer) {
    var buffer bytes.Buffer
    view.RenderPartial(&buffer, "server_error.tmpl", nil)
    view.RenderLayout(w, &view.RenderInfo{
        Yield: view.HTML(buffer.String()),
        Title: "Oh. Sorry about that.",
    })
}

func notFound(w io.Writer) {
    var buffer bytes.Buffer
    view.RenderPartial(&buffer, "not_found.tmpl", nil)
    view.RenderLayout(w, &view.RenderInfo{
        Yield: view.HTML(buffer.String()),
        Title: "Not Found",
    })
}

func redirectHandler(req *web.Request) {
    url := req.URL
    url.Host = config.CanonicalHost
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
        Register("/sitemap.xml<gzip:(\\.gz)?>", "GET", sitemapHandler).
        Register("/archive/<archive:(full|category|month)>", "GET", archiveHandler).
        Register("/<year:\\d{4}>/<month:\\d{2}>", "GET", monthlyHandler).
        Register("/category/<category>", "GET", categoryHandler).
        Register("/<year:\\d{4}>/<month:\\d{2}>/<day:\\d{2}>/<slug>", "GET", permalinkHandler).
        Register("/tag/<tag>", "GET", tagHandler).
        Register("/<slug:\\w+>", "GET", pageHandler).
        Register("/<path:.*>", "GET", web.DirectoryHandler("public", nil))

    redirector := web.NewRouter().
        Register("/<splat:.*>", "GET", redirectHandler)

    hostRouter := web.NewHostRouter(redirector).
        Register(config.CanonicalHost, router)

    listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.Port))
    if err != nil {
        logger.Fatalf("Failed to listen: %s", err)
    }
    defer listener.Close()
    server := &server.Server{
        Listener: listener,
        Handler:  hostRouter,
        Logger:   server.LoggerFunc(ShortLogger),
    }
    logger.Printf("verboselogging is starting on 0.0.0.0:%d", config.Port)
    err = server.Serve()
    if err != nil {
        logger.Fatalf("Failed to serve: %s", err)
    }
}
