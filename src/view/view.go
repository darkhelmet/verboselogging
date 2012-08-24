package view

import (
    "encoding/json"
    "fmt"
    "github.com/darkhelmet/env"
    T "html/template"
    "io"
    "io/ioutil"
    "log"
    "os"
)

var (
    templates     *T.Template
    port          = env.IntDefault("PORT", 8080)
    canonicalHost = env.StringDefaultF("CANONICAL_HOST", func() string { return fmt.Sprintf("localhost:%d", port) })
    assetHost     = env.StringDefaultF("ASSET_HOST", func() string { return fmt.Sprintf("http://%s", canonicalHost) })
    logger        = log.New(os.Stdout, "[view] ", env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds))
    middot        = T.HTML("&middot;")
    HTML          = func(s string) T.HTML { return T.HTML(s) }
    pageLinks     = []PageLink{
        PageLink{Name: "Home", Path: "/", Icon: "H", Header: true, Footer: true, After: middot},
        PageLink{Name: "About", Path: "/about", Icon: "A", Header: true, Footer: true, After: middot},
        PageLink{Name: "Talks", Path: "/talks", Icon: "E", Header: true, Footer: true, After: middot},
        PageLink{Name: "Projects", Path: "/projects", Icon: "P", Header: true, Footer: true, After: middot},
        PageLink{Name: "Contact", Path: "/contact", Icon: "C", Header: true, Footer: true, After: middot},
        PageLink{Name: "Disclaimer", Path: "/disclaimer", Icon: "D", Header: true, Footer: true, After: middot},
        PageLink{Name: "Sitemap", Path: "/sitemap.xml", Footer: true},
    }
    assets = make(map[string]string)
)

type PageLink struct {
    Name, Path, Class, Icon string
    After                   T.HTML
    Header, Footer          bool
}

type TemplateData struct {
    Yield                         interface{}
    Title, Description, Canonical string
    PageLinks                     []PageLink
}

func setupAssets() {
    manifest := make(map[string]interface{})
    data, err := ioutil.ReadFile("public/assets/manifest.json")
    if err != nil {
        logger.Fatalf("failed to read asset manifest file: %s", err)
    }
    err = json.Unmarshal(data, &manifest)
    if err != nil {
        logger.Fatalf("failed parsing manifest file: %s", err)
    }
    pairs := manifest["assets"].(map[string]interface{})
    for key, path := range pairs {
        assets[key] = path.(string)
    }
}

func init() {
    templates = T.Must(T.New("funcs").Funcs(T.FuncMap{
        "ArchivePath":    archivePath,
        "FontTag":        fontTag,
        "StylesheetPath": stylesheetPath,
        "JavascriptPath": javascriptPath,
        "ImagePath":      imagePath,
        "AssetPath":      assetPath,
        "Url":            urlFunction,
    }).ParseGlob("views/*.tmpl"))
    setupAssets()
}

func urlFunction(path string) string {
    return fmt.Sprintf("http://%s%s", canonicalHost, path)
}

func assetPath(name string) string {
    return fmt.Sprintf("%s/assets/%s", assetHost, assets[name])
}

func stylesheetPath(name string) string {
    return assetPath(fmt.Sprintf("stylesheets/%s.css", name))
}

func imagePath(name string) string {
    return assetPath(fmt.Sprintf("images/%s", name))
}

func javascriptPath(name string) string {
    return assetPath(fmt.Sprintf("javascripts/%s.js", name))
}

func archivePath(name string) string {
    return fmt.Sprintf("/archive/%s", name)
}

func fontTag(family string) T.HTML {
    return T.HTML(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="http://fonts.googleapis.com/css?family=%s">`, family))
}

func RenderLayout(w io.Writer, yield interface{}, path, title, description string) {
    if title == "" {
        title = "Verbose Logging"
    } else {
        title = fmt.Sprintf("%s | Verbose Logging", title)
    }
    err := templates.ExecuteTemplate(w, "layout.tmpl", TemplateData{
        Yield:       yield,
        Title:       title,
        Description: description,
        PageLinks:   pageLinks,
        Canonical:   path,
    })
    if err != nil {
        logger.Printf("error rendering template: %s", err)
    }
}

func RenderPartial(w io.Writer, name string, data interface{}) {
    err := templates.ExecuteTemplate(w, name, data)
    if err != nil {
        logger.Printf("error rendering partial: %s", err)
    }
}
