package main

import (
    "fmt"
    "github.com/darkhelmet/env"
    "io"
    "log"
    "net"
    "os"
    "vendor/github.com/garyburd/twister/server"
    "vendor/github.com/garyburd/twister/web"
)

var (
    port          = env.IntDefault("PORT", 8080)
    canonicalHost = env.StringDefaultF("CANONICAL_HOST", func() string { return fmt.Sprintf("localhost:%d", port) })
    logger        = log.New(os.Stdout, "[server] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
)

func rootHandler(req *web.Request) {
    w := req.Respond(web.StatusOK)
    io.WriteString(w, "rootHandler")
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

/*
      root            /                                     posts#main
opensearch GET        /opensearch.xml(.:format)             application#opensearch {:format=>:xml}
    search GET        /search(.:format)                     posts#search
      feed GET        /feed(.:format)                       posts#feed {:format=>:xml}
   sitemap GET        /sitemap.:format                      posts#sitemap {:format=>:xml}
   archive GET        /archive/:archive(.:format)           posts#archive {:archive=>/full|category|month/}
   monthly GET        /:year/:month(.:format)               posts#monthly {:year=>/\d{4}/, :month=>/\d{2}/}
  category GET        /category/:category(.:format)         posts#category
 permalink GET        /:year/:month/:day/:slug(.:format)    posts#permalink {:year=>/\d{4}/, :month=>/\d{2}/, :day=>/\d{2}/}
       tag GET        /tag/:tag(.:format)                   posts#tag
      page GET        /:page(.:format)                      pages#show
           GET        /*not_found(.:format)                 application#render_404
*/

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
        Register("/<page>", "GET", pageHandler)

    redirector := web.NewRouter().
        Register("/", "GET", redirectHandler).
        Register("/<splat>", "GET", redirectHandler)

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
