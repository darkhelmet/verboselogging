package view

import (
    "config"
    "encoding/json"
    "fmt"
    T "html/template"
    "io"
    "io/ioutil"
    "log"
    "os"
)

var (
    templates *T.Template
    logger    = log.New(os.Stdout, "[view] ", config.LogFlags)
    middot    = T.HTML("&middot;")
    HTML      = func(s string) T.HTML { return T.HTML(s) }
    pageLinks = []PageLink{
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
    Yield                                   interface{}
    Title, Description, Canonical           string
    SiteTitle, SiteDescription, SiteContact string
    PageLinks                               []PageLink
}

type RenderInfo struct {
    Yield                         interface{}
    Title, Description, Canonical string
}

func setupAssets() {
    data, err := ioutil.ReadFile("public/assets/manifest.json")
    if err != nil {
        logger.Fatalf("failed to read asset manifest file: %s", err)
    }
    manifest := make(map[string]interface{})
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
        "AssetPath": assetPath,
        "ArchivePath": func(name string) string {
            return fmt.Sprintf("/archive/%s", name)
        },
        "FontTag": func(family string) T.HTML {
            return T.HTML(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="http://fonts.googleapis.com/css?family=%s">`, family))
        },
        "StylesheetPath": func(name string) string {
            return assetPath(fmt.Sprintf("stylesheets/%s.css", name))
        },
        "JavascriptPath": func(name string) string {
            return assetPath(fmt.Sprintf("javascripts/%s.js", name))
        },
        "ImagePath": func(name string) string {
            return assetPath(fmt.Sprintf("images/%s", name))
        },
        "CanonicalUrl": func(path string) string {
            return fmt.Sprintf("http://%s%s", config.CanonicalHost, path)
        },
    }).ParseGlob("views/*.tmpl"))
    setupAssets()
}

func assetPath(name string) string {
    return fmt.Sprintf("%s/assets/%s", config.AssetHost, assets[name])
}

func RenderLayout(w io.Writer, data *RenderInfo) {
    err := templates.ExecuteTemplate(w, "layout.tmpl", TemplateData{
        Yield:           data.Yield,
        Title:           data.Title,
        Description:     data.Description,
        Canonical:       data.Canonical,
        SiteTitle:       config.SiteTitle,
        SiteDescription: config.SiteDescription,
        SiteContact:     config.SiteContact,
        PageLinks:       pageLinks,
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
