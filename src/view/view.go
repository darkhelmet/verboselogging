package view

import (
    "fmt"
    "github.com/darkhelmet/env"
    T "html/template"
    "io"
    "log"
    "os"
)

var (
    templates *T.Template
    logger    = log.New(os.Stdout, "[view] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
    middot    = T.HTML("&middot;")
    pageLinks = []PageLink{
        PageLink{Name: "Home", Path: "/", Icon: "H", Header: true, Footer: true, After: middot},
        PageLink{Name: "About", Path: "/about", Icon: "A", Header: true, Footer: true, After: middot},
        PageLink{Name: "Talks", Path: "/talks", Icon: "E", Header: true, Footer: true, After: middot},
        PageLink{Name: "Projects", Path: "/projects", Icon: "P", Header: true, Footer: true, After: middot},
        PageLink{Name: "Contact", Path: "/contact", Icon: "C", Header: true, Footer: true, After: middot},
        PageLink{Name: "Disclaimer", Path: "/disclaimer", Icon: "D", Header: true, Footer: true, After: middot},
        PageLink{Name: "Sitemap", Path: "/sitemap.xml", Footer: true},
    }
)

type PageLink struct {
    Name, Path, Class, Icon string
    After                   T.HTML
    Header, Footer          bool
}

type TemplateData struct {
    Title, Yield, Description, Canonical string
    PageLinks                            []PageLink
}

func init() {
    templates = T.Must(T.New("funcs").Funcs(T.FuncMap{
        "ArchivePath": archivePath,
        "FontTag":     fontTag,
    }).ParseGlob("views/*.tmpl"))
}

func archivePath(name string) string {
    return fmt.Sprintf("/archive/%s", name)
}

func fontTag(family string) T.HTML {
    return T.HTML(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="http://fonts.googleapis.com/css?family=%s">`, family))
}

func RenderLayout(w io.Writer, yield, title, description string) {
    err := templates.ExecuteTemplate(w, "layout.tmpl", TemplateData{
        Yield:       yield,
        Title:       title,
        Description: description,
        PageLinks:   pageLinks,
    })
    if err != nil {
        logger.Printf("error rendering template: %s", err)
    }
}
