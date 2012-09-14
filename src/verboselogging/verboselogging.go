package main

import (
    "config"
    "fmt"
    "github.com/darkhelmet/env"
    "log"
    "net"
    "os"
    Page "page"
    Post "post"
    "regexp"
    "strings"
    "time"
    "vendor/github.com/garyburd/twister/server"
    "vendor/github.com/garyburd/twister/web"
    "vendor/github.com/tobi/airbrake-go"
    "view"
)

var (
    logger        = log.New(os.Stdout, "[server] ", config.LogFlags)
    feedburner    = regexp.MustCompile("(?i)feedburner")
    feedburnerUrl = "http://feeds.feedburner.com/VerboseLogging"
)

func init() {
    airbrake.ApiKey = env.StringDefault("AIRBRAKE_TOKEN", "")
}

func rootHandler(req *web.Request, r Responder) {
    posts, err := Post.FindLatest(6)
    if err != nil {
        logger.Printf("failed finding latest posts: %s", err)
        serverError(r, err)
    } else {
        w := r.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        view.RenderLayout(w, &view.RenderInfo{
            PostPreview:  posts,
            Canonical:    "/",
            ArchiveLinks: true,
            Description:  config.SiteDescription,
        })
    }
}

func opensearchHandler(req *web.Request, r Responder) {
    w := r.Respond(web.StatusOK, web.HeaderContentType, "application/xml; charset=utf-8")
    view.RenderPartial(w, "opensearch.tmpl", nil)
}

func searchHandler(req *web.Request, r Responder) {
    query := req.Param.Get("query")
    posts, err := Post.Search(query)
    if err != nil {
        logger.Printf("failed finding posts with query %#v: %s", query, err)
        serverError(r, err)
    } else {
        w := r.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        title := fmt.Sprintf("Search results for %#v", query)
        view.RenderLayout(w, &view.RenderInfo{
            PostPreview:  posts,
            Title:        title,
            PageTitle:    title,
            ArchiveLinks: true,
        })
    }
}

func feedHandler(req *web.Request, r Responder) {
    if !feedburner.Match([]byte(req.Header.Get(web.HeaderUserAgent))) {
        // Not Feedburner
        if "" == req.Param.Get("no_fb") {
            // And nothing saying to ignore
            r.Respond(web.StatusMovedPermanently, web.HeaderLocation, feedburnerUrl)
            return
        }
    }

    posts, err := Post.FindLatest(10)
    if err != nil {
        logger.Printf("failed getting posts for feed: %s", err)
        serverError(r, err)
    } else {
        w := r.Respond(web.StatusOK, web.HeaderContentType, "application/rss+xml; charset=utf-8")
        view.RenderPartial(w, "feed.tmpl", &view.RenderInfo{Post: posts})
    }
}

func sitemapHandler(req *web.Request, r Responder) {
    posts, err := Post.FindForSitemap()
    if err != nil {
        logger.Printf("failed getting posts for sitemap: %s", err)
        serverError(r, err)
    } else {
        w := r.Respond(web.StatusOK, web.HeaderContentType, "application/xml; charset=utf-8")
        view.RenderPartial(w, "sitemap.tmpl", &view.RenderInfo{Post: posts})
    }
}

func fullArchiveHandler(req *web.Request, r Responder) {
    posts, err := Post.FindForArchive()
    if err != nil {
        logger.Printf("failed getting posts for full archive: %s", err)
        serverError(r, err)
    } else {
        w := r.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        title := "Full archives"
        view.RenderLayout(w, &view.RenderInfo{
            FullArchive:  posts,
            Description:  title,
            Title:        title,
            ArchiveLinks: true,
            Canonical:    req.URL.Path,
        })
    }
}

func categoryArchiveHandler(req *web.Request, r Responder) {
    posts, err := Post.FindForArchive()
    if err != nil {
        logger.Printf("failed getting posts for category archive: %s", err)
        serverError(r, err)
    } else {
        grouped := make(map[string][]*Post.Post)
        for _, post := range posts {
            key := post.Category
            grouped[key] = append(grouped[key], post)
        }

        w := r.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        view.RenderLayout(w, &view.RenderInfo{
            CategoryArchive: grouped,
            Description:     "Archives by category",
            Title:           "Category archives",
            ArchiveLinks:    true,
            Canonical:       req.URL.Path,
        })
    }
}

func monthlyArchiveHandler(req *web.Request, r Responder) {
    posts, err := Post.FindForArchive()
    if err != nil {
        logger.Printf("failed getting posts for monthly archive: %s", err)
        serverError(r, err)
    } else {
        grouped := make(map[int64][]*Post.Post)
        for _, post := range posts {
            t := post.PublishedOn.In(config.TimeZone)
            key := -time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local).Unix()
            grouped[key] = append(grouped[key], post)
        }

        w := r.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        view.RenderLayout(w, &view.RenderInfo{
            MonthlyArchive: grouped,
            Description:    "Archives by month",
            Title:          "Monthly archives",
            ArchiveLinks:   true,
            Canonical:      req.URL.Path,
        })
    }
}

func monthlyHandler(req *web.Request, r Responder) {
    year, month := req.URLParam["year"], req.URLParam["month"]
    posts, err := Post.FindByMonth(year, month)
    if err != nil {
        logger.Printf("failed finding posts in month %#v of %#v: %s", month, year, err)
        serverError(r, err)
    } else {
        w := r.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        title := fmt.Sprintf("Archives for %s-%s", month, year)
        view.RenderLayout(w, &view.RenderInfo{
            PostPreview:  posts,
            Title:        title,
            Canonical:    req.URL.Path,
            ArchiveLinks: true,
            Description:  title,
        })
    }
}

func categoryHandler(req *web.Request, r Responder) {
    category := req.URLParam["category"]
    posts, err := Post.FindByCategory(category)
    if err != nil {
        logger.Printf("failed finding posts with category %#v: %s", category, err)
        serverError(r, err)
    } else {
        w := r.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        category = strings.Title(category)
        title := fmt.Sprintf("%s Articles", category)
        view.RenderLayout(w, &view.RenderInfo{
            PostPreview: posts,
            Title:       title,
            PageTitle:   title,
            Canonical:   req.URL.Path,
            Description: fmt.Sprintf("Articles in the %s category", category),
        })
    }
}

func permalinkHandler(req *web.Request, r Responder) {
    slug := req.URLParam["slug"]
    year, month, day := req.URLParam["year"], req.URLParam["month"], req.URLParam["day"]
    post, err := Post.FindByPermalink(year, month, day, slug)
    if err != nil {
        switch err.(type) {
        case Post.NotFound:
            notFound(r)
        default:
            logger.Printf("failed finding post with year(%#v) month(%#v) day(%#v) slug(%#v): %s (%T)", year, month, day, slug, err, err)
            serverError(r, err)
        }
    } else {
        w := r.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        view.RenderLayout(w, &view.RenderInfo{
            Post:        post,
            Title:       post.Title,
            Canonical:   post.Canonical(),
            Description: post.Description,
        })
    }
}

func tagHandler(req *web.Request, r Responder) {
    tag := req.URLParam["tag"]
    posts, err := Post.FindByTag(tag)
    if err != nil {
        logger.Printf("failed finding posts with tag %#v: %s", tag, err)
        serverError(r, err)
    } else {
        w := r.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        title := fmt.Sprintf("Articles tagged with %#v", tag)
        view.RenderLayout(w, &view.RenderInfo{
            PostPreview: posts,
            Title:       title,
            PageTitle:   title,
            Canonical:   req.URL.Path,
            Description: fmt.Sprintf("Articles with the %#v tag", tag),
        })
    }
}

func pageHandler(req *web.Request, r Responder) {
    slug := req.URLParam["slug"]
    page, err := Page.FindBySlug(slug)
    if err != nil {
        switch err.(type) {
        case Page.NotFound:
            notFound(r)
        default:
            logger.Printf("failed finding page with slug %#v: %s (%T)", slug, err, err)
            serverError(r, err)
        }
    } else {
        w := r.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        view.RenderLayout(w, &view.RenderInfo{
            Page:        view.HTML(page.BodyHtml),
            Title:       page.Title,
            Canonical:   page.Canonical(),
            Description: page.Description,
        })
    }
}

func serverError(r Responder, err error) {
    airbrake.Error(err, nil)
    w := r.Respond(web.StatusInternalServerError, web.HeaderContentType, "text/html; charset=utf-8")
    view.RenderLayout(w, &view.RenderInfo{
        Error: true,
        Title: "Oh. Sorry about that.",
    })
}

func notFound(r Responder) {
    w := r.Respond(web.StatusNotFound, web.HeaderContentType, "text/html; charset=utf-8")
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
        Register("/", "GET", Use(Gzip(rootHandler))).
        Register("/opensearch.xml", "GET", Use(Gzip(opensearchHandler))).
        Register("/search", "GET", Use(Gzip(searchHandler))).
        Register("/feed", "GET", Use(Gzip(feedHandler))).
        Register("/sitemap.xml<gzip:(\\.gz)?>", "GET", Use(Gzip(sitemapHandler))).
        Register("/archive/full", "GET", Use(Gzip(fullArchiveHandler))).
        Register("/archive/category", "GET", Use(Gzip(categoryArchiveHandler))).
        Register("/archive/month", "GET", Use(Gzip(monthlyArchiveHandler))).
        Register("/<year:\\d{4}>/<month:\\d{2}>", "GET", Use(Gzip(monthlyHandler))).
        Register("/category/<category>", "GET", Use(Gzip(categoryHandler))).
        Register("/<year:\\d{4}>/<month:\\d{2}>/<day:\\d{2}>/<slug>", "GET", Use(Gzip(permalinkHandler))).
        Register("/tag/<tag>", "GET", Use(Gzip(tagHandler))).
        Register("/<slug:\\w+>", "GET", Use(Gzip(pageHandler))).
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
