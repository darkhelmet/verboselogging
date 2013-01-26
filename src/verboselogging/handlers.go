package verboselogging

import (
    "config"
    "fmt"
    "github.com/darkhelmet/blargh/errors"
    "github.com/darkhelmet/blargh/post"
    "github.com/darkhelmet/webutil"
    "log"
    "net/http"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"
    "vendor/github.com/garyburd/twister/adapter"
    "vendor/github.com/garyburd/twister/web"
    "view"
)

var (
    logger        = log.New(os.Stdout, "[verboselogging] ", config.LogFlags)
    feedburner    = regexp.MustCompile("(?i)feedburner")
    feedburnerUrl = "http://feeds.feedburner.com/VerboseLogging"
    posts         = NewRepo("posts")
    pages         = NewRepo("pages")
)

func rootHandler(req *web.Request) {
    posts, err := posts.FindLatest(6)
    if err != nil {
        logger.Printf("failed finding latest posts: %s", err)
        serverError(req, err)
    } else {
        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        view.RenderLayout(w, &view.RenderInfo{
            PostPreview:  posts,
            Canonical:    "/",
            ArchiveLinks: true,
            Description:  config.SiteDescription,
        })
    }
}

func opensearchHandler(req *web.Request) {
    w := req.Respond(web.StatusOK, web.HeaderContentType, "application/xml; charset=utf-8")
    view.RenderPartial(w, "opensearch.tmpl", nil)
}

func searchHandler(req *web.Request) {
    query := req.Param.Get("query")
    posts, err := posts.Search(query)
    if err != nil {
        logger.Printf("failed finding posts with query %#v: %s", query, err)
        serverError(req, err)
    } else {
        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        title := fmt.Sprintf("Search results for %#v", query)
        view.RenderLayout(w, &view.RenderInfo{
            PostPreview:  posts,
            Title:        title,
            PageTitle:    title,
            ArchiveLinks: true,
        })
    }
}

func feedHandler(req *web.Request) {
    if !feedburner.Match([]byte(req.Header.Get(web.HeaderUserAgent))) {
        // Not Feedburner
        if "" == req.Param.Get("no_fb") {
            // And nothing saying to ignore
            req.Respond(web.StatusMovedPermanently, web.HeaderLocation, feedburnerUrl)
            return
        }
    }

    posts, err := posts.FindLatest(10)
    if err != nil {
        logger.Printf("failed getting posts for feed: %s", err)
        serverError(req, err)
    } else {
        w := req.Respond(web.StatusOK, web.HeaderContentType, "application/rss+xml; charset=utf-8")
        view.RenderPartial(w, "feed.tmpl", &view.RenderInfo{
            Post:            posts,
            SiteTitle:       config.SiteTitle,
            SiteDescription: config.SiteDescription,
            SiteContact:     config.SiteContact,
            SiteAuthor:      config.SiteAuthor,
        })
    }
}

func sitemapHandler(req *web.Request) {
    posts, err := posts.FindLatest(posts.Len())
    if err != nil {
        logger.Printf("failed getting posts for sitemap: %s", err)
        serverError(req, err)
    } else {
        w := req.Respond(web.StatusOK, web.HeaderContentType, "application/xml; charset=utf-8")
        view.RenderPartial(w, "sitemap.tmpl", &view.RenderInfo{Post: posts})
    }
}

func fullArchiveHandler(req *web.Request) {
    posts, err := posts.FindLatest(posts.Len())
    if err != nil {
        logger.Printf("failed getting posts for full archive: %s", err)
        serverError(req, err)
    } else {
        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
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

func categoryArchiveHandler(req *web.Request) {
    posts, err := posts.FindLatest(posts.Len())
    if err != nil {
        logger.Printf("failed getting posts for category archive: %s", err)
        serverError(req, err)
    } else {
        grouped := make(map[string][]*post.Post)
        for _, post := range posts {
            key := post.Category
            grouped[key] = append(grouped[key], post)
        }

        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        view.RenderLayout(w, &view.RenderInfo{
            CategoryArchive: grouped,
            Description:     "Archives by category",
            Title:           "Category archives",
            ArchiveLinks:    true,
            Canonical:       req.URL.Path,
        })
    }
}

func monthlyArchiveHandler(req *web.Request) {
    posts, err := posts.FindLatest(posts.Len())
    if err != nil {
        logger.Printf("failed getting posts for monthly archive: %s", err)
        serverError(req, err)
    } else {
        grouped := make(map[int64][]*post.Post)
        for _, post := range posts {
            t := post.PublishedOn
            key := -time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local).Unix()
            grouped[key] = append(grouped[key], post)
        }

        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        view.RenderLayout(w, &view.RenderInfo{
            MonthlyArchive: grouped,
            Description:    "Archives by month",
            Title:          "Monthly archives",
            ArchiveLinks:   true,
            Canonical:      req.URL.Path,
        })
    }
}

func monthlyHandler(req *web.Request) {
    year, month := req.URLParam["year"], req.URLParam["month"]
    y, _ := strconv.Atoi(year)
    m, _ := strconv.Atoi(month)
    posts, err := posts.FindByMonth(y, time.Month(m))
    if err != nil {
        logger.Printf("failed finding posts in month %#v of %#v: %s", month, year, err)
        serverError(req, err)
    } else {
        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
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

func categoryHandler(req *web.Request) {
    category := req.URLParam["category"]
    posts, err := posts.FindByCategory(category)
    if err != nil {
        logger.Printf("failed finding posts with category %#v: %s", category, err)
        serverError(req, err)
    } else {
        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
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

func permalinkHandler(req *web.Request) {
    slug := req.URLParam["slug"]
    year, month, day := req.URLParam["year"], req.URLParam["month"], req.URLParam["day"]
    y, _ := strconv.Atoi(year)
    m, _ := strconv.Atoi(month)
    d, _ := strconv.Atoi(day)
    post, err := posts.FindByPermalink(y, time.Month(m), d, slug)
    if err != nil {
        switch err.(type) {
        case errors.NotFound:
            notFound(req)
        default:
            logger.Printf("failed finding post with year(%#v) month(%#v) day(%#v) slug(%#v): %s (%T)", year, month, day, slug, err, err)
            serverError(req, err)
        }
    } else {
        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        view.RenderLayout(w, &view.RenderInfo{
            Post:        post,
            Title:       post.Title,
            Canonical:   view.PostCanonical(post),
            Description: post.Description,
        })
    }
}

func tagHandler(req *web.Request) {
    tag := req.URLParam["tag"]
    posts, err := posts.FindByTag(tag)
    if err != nil {
        logger.Printf("failed finding posts with tag %#v: %s", tag, err)
        serverError(req, err)
    } else {
        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
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

func pageHandler(req *web.Request) {
    slug := req.URLParam["slug"]
    page, err := pages.FindBySlug(slug)
    if err != nil {
        switch err.(type) {
        case errors.NotFound:
            notFound(req)
        default:
            logger.Printf("failed finding page with slug %#v: %s (%T)", slug, err, err)
            serverError(req, err)
        }
    } else {
        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/html; charset=utf-8")
        view.RenderLayout(w, &view.RenderInfo{
            Page:        page,
            Title:       page.Title,
            Canonical:   view.PageCanonical(page),
            Description: page.Description,
        })
    }
}

func serverError(req *web.Request, err error) {
    w := req.Respond(web.StatusInternalServerError, web.HeaderContentType, "text/html; charset=utf-8")
    view.RenderLayout(w, &view.RenderInfo{
        Error: true,
        Title: "Oh. Sorry about that.",
    })
}

func notFound(req *web.Request) {
    w := req.Respond(web.StatusNotFound, web.HeaderContentType, "text/html; charset=utf-8")
    view.RenderLayout(w, &view.RenderInfo{
        NotFound: true,
        Title:    "Not Found",
    })
}

func SetupHandler() http.Handler {
    staticOptions := &web.ServeFileOptions{
        Header: web.Header{
            web.HeaderCacheControl:              {"public, max-age=31536000"},
            web.HeaderAccessControllAllowOrigin: {"*"},
        },
    }

    router := web.NewRouter().
        Register("/", "GET", rootHandler).
        Register("/opensearch.xml", "GET", opensearchHandler).
        Register("/search", "GET", searchHandler).
        Register("/feed", "GET", feedHandler).
        Register("/sitemap.xml<gzip:(\\.gz)?>", "GET", sitemapHandler).
        Register("/archive/full", "GET", fullArchiveHandler).
        Register("/archive/category", "GET", categoryArchiveHandler).
        Register("/archive/month", "GET", monthlyArchiveHandler).
        Register("/<year:\\d{4}>/<month:\\d{2}>", "GET", monthlyHandler).
        Register("/category/<category>", "GET", categoryHandler).
        Register("/<year:\\d{4}>/<month:\\d{2}>/<day:\\d{2}>/<slug>", "GET", permalinkHandler).
        Register("/tag/<tag>", "GET", tagHandler).
        Register("/<slug:\\w+>", "GET", pageHandler).
        Register("/<path:.*>", "GET", web.DirectoryHandler("public", staticOptions))

    var handler http.Handler = adapter.HTTPHandler{router}
    handler = webutil.GzipHandler{handler}
    handler = webutil.LoggerHandler{handler, logger}
    handler = webutil.HerokuHandler{handler, logger}
    handler = webutil.CanonicalHostHandler{handler, config.CanonicalHost, "http"}
    handler = webutil.EnsureRequestBodyClosedHandler{handler}
    return handler
}
