package view

import (
    "config"
    "crypto/md5"
    "encoding/json"
    "fmt"
    "github.com/darkhelmet/blargh/post"
    T "html/template"
    "io"
    "io/ioutil"
    "log"
    "os"
    "strings"
    "time"
    "unicode"
    "unicode/utf8"
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

type Formatter interface {
    Format(string) string
}

type TimeZoner interface {
    In(*time.Location) time.Time
}

type PageLink struct {
    Name, Path, Class, Icon string
    After                   T.HTML
    Header, Footer          bool
}

type RenderInfo struct {
    Page                                               interface{}
    Title, PageTitle, Description, Canonical, Gravatar string
    Error, NotFound, ArchiveLinks                      bool

    SiteTitle, SiteDescription, SiteContact, SiteAuthor             string
    PageLinks                                                       []PageLink
    PostPreview, Post, FullArchive, CategoryArchive, MonthlyArchive interface{}
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
        "Time": func(s int64) time.Time {
            return time.Unix(-s, 0)
        },
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
        "ISO8601": func(t Formatter) string {
            return t.Format(time.RFC3339)
        },
        "RFC822": func(t Formatter) string {
            return t.Format(time.RFC822)
        },
        "DisplayTime": func(t Formatter) string {
            return t.Format("02 Jan 2006 15:04 MST")
        },
        "Gravatar": func(email string) string {
            email = strings.TrimFunc(email, unicode.IsSpace)
            email = strings.ToLower(email)
            hash := md5.New()
            io.WriteString(hash, email)
            return fmt.Sprintf("http://www.gravatar.com/avatar/%x.png", hash.Sum(nil))
        },
        "CategoryPath": func(i interface{}) string {
            switch thing := i.(type) {
            case *post.Post:
                return fmt.Sprintf("/category/%s", thing.Category)
            case string:
                return fmt.Sprintf("/category/%s", thing)
            default:
                panic("YOU SHALL NOT PASS!!!")
            }
            panic("not reachable")
        },
        "UTC": func(t TimeZoner) time.Time {
            return t.In(time.UTC)
        },
        "MonthlyPath": func(t Formatter) string {
            return t.Format("/2006/01")
        },
        "TagPath": func(tag string) string {
            return fmt.Sprintf("/tag/%s", tag)
        },
        "Truncate": func(length int, s string) string {
            if length < utf8.RuneCountInString(s) {
                trimmed := []rune(s)[0:length]
                trimmed[length-1] = '…'
                return string(trimmed)
            }
            return s
        },
        "Titleize":      strings.Title,
        "Safe":          HTML,
        "PostCanonical": PostCanonical,
        "PageCanonical": PageCanonical,
    }).ParseGlob("views/*.tmpl"))
    setupAssets()
}

func assetPath(name string) string {
    return fmt.Sprintf("%s/assets/%s", config.AssetHost, assets[name])
}

func RenderLayout(w io.Writer, data *RenderInfo) {
    data.SiteTitle = config.SiteTitle
    data.SiteDescription = config.SiteDescription
    data.SiteContact = config.SiteContact
    data.SiteAuthor = config.SiteAuthor
    data.PageLinks = pageLinks
    err := templates.ExecuteTemplate(w, "layout.tmpl", data)
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

func PostCanonical(p *post.Post) string {
    return fmt.Sprintf("/%s/%s", p.PublishedOn.Format("2006/01/02"), p.Slug())
}

func PageCanonical(p *post.Post) string {
    return "/" + p.Slug()
}
