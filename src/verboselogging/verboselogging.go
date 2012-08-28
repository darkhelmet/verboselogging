package main

import (
    "compress/gzip"
    "config"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    Page "page"
    Post "post"
    "strings"
    "vendor/github.com/garyburd/twister/server"
    "vendor/github.com/garyburd/twister/web"
    "view"
)

var (
    logger = log.New(os.Stdout, "[server] ", config.LogFlags)
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
    posts, err := Post.FindLatest(6)
    if err != nil {
        logger.Printf("failed finding latest posts: %s", err)
        withGzip(req, web.StatusInternalServerError, "text/html; charset=utf-8", serverError)
    } else {
        withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
            view.RenderLayout(w, &view.RenderInfo{
                PostPreview:  posts,
                Canonical:    "/",
                ArchiveLinks: true,
                Description:  config.SiteDescription,
            })
        })
    }
}

func opensearchHandler(req *web.Request) {
    withGzip(req, web.StatusOK, "application/xml; charset=utf-8", func(w io.Writer) {
        view.RenderPartial(w, "opensearch.tmpl", nil)
    })
}

func searchHandler(req *web.Request) {
    query := req.Param.Get("query")
    posts, err := Post.Search(query)
    if err != nil {
        logger.Printf("failed finding posts with query %#v: %s", query, err)
        withGzip(req, web.StatusInternalServerError, "text/html; charset=utf-8", serverError)
    } else {
        withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
            title := fmt.Sprintf("Search results for %#v", query)
            view.RenderLayout(w, &view.RenderInfo{
                PostPreview: posts,
                Title:       title,
                PageTitle:   title,
                // Canonical:   req.URL.Path,
            })
        })
    }
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
    category := req.URLParam["category"]
    posts, err := Post.FindByCategory(category)
    if err != nil {
        logger.Printf("failed finding posts with category %#v: %s", category, err)
        withGzip(req, web.StatusInternalServerError, "text/html; charset=utf-8", serverError)
    } else {
        withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
            category = strings.Title(category)
            title := fmt.Sprintf("%s Articles", category)
            view.RenderLayout(w, &view.RenderInfo{
                PostPreview: posts,
                Title:       title,
                PageTitle:   title,
                Canonical:   req.URL.Path,
                Description: fmt.Sprintf("Articles in the %s category", category),
            })
        })
    }
}

func permalinkHandler(req *web.Request) {
    slug := req.URLParam["slug"]
    year, month, day := req.URLParam["year"], req.URLParam["month"], req.URLParam["day"]
    post, err := Post.FindByPermalink(year, month, day, slug)
    if err != nil {
        switch err.(type) {
        case Post.NotFound:
            withGzip(req, web.StatusNotFound, "text/html; charset=utf-8", notFound)
        default:
            logger.Printf("failed finding post with year(%#v) month(%#v) day(%#v) slug(%#v): %s (%T)", year, month, day, slug, err, err)
            withGzip(req, web.StatusInternalServerError, "text/html; charset=utf-8", serverError)
        }
    } else {
        withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
            view.RenderLayout(w, &view.RenderInfo{
                Post:        post,
                Title:       post.Title,
                Canonical:   post.Canonical(),
                Description: post.Description,
            })
        })
    }
}

func tagHandler(req *web.Request) {
    tag := req.URLParam["tag"]
    posts, err := Post.FindByTag(tag)
    if err != nil {
        logger.Printf("failed finding posts with tag %#v: %s", tag, err)
        withGzip(req, web.StatusInternalServerError, "text/html; charset=utf-8", serverError)
    } else {
        withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
            title := fmt.Sprintf("Articles tagged with %#v", tag)
            view.RenderLayout(w, &view.RenderInfo{
                PostPreview: posts,
                Title:       title,
                PageTitle:   title,
                Canonical:   req.URL.Path,
            })
        })
    }
}

func pageHandler(req *web.Request) {
    slug := req.URLParam["slug"]
    page, err := Page.FindBySlug(slug)
    if err != nil {
        switch err.(type) {
        case Page.NotFound:
            withGzip(req, web.StatusNotFound, "text/html; charset=utf-8", notFound)
        default:
            logger.Printf("failed finding page with slug %#v: %s (%T)", slug, err, err)
            withGzip(req, web.StatusInternalServerError, "text/html; charset=utf-8", serverError)
        }
    } else {
        withGzip(req, web.StatusOK, "text/html; charset=utf-8", func(w io.Writer) {
            view.RenderLayout(w, &view.RenderInfo{
                Page:        view.HTML(page.BodyHtml),
                Title:       page.Title,
                Canonical:   page.Canonical(),
                Description: config.SiteDescription,
            })
        })
    }
}

func serverError(w io.Writer) {
    view.RenderLayout(w, &view.RenderInfo{
        Error: true,
        Title: "Oh. Sorry about that.",
    })
}

func notFound(w io.Writer) {
    view.RenderLayout(w, &view.RenderInfo{
        NotFound: true,
        Title:    "Not Found",
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
    staticOptions := &web.ServeFileOptions{
        Header: web.Header{
            web.HeaderCacheControl: {"public, max-age=31536000"},
        },
    }
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
        Register("/<path:.*>", "GET", web.DirectoryHandler("public", staticOptions))

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
