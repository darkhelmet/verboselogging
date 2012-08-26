package view

import (
    "config"
    "fmt"
    T "html/template"
    "post"
)

func FuncMap() T.FuncMap {
    return T.FuncMap{
        "PostPermalinkPath": func(p *post.Post) string {
            return fmt.Sprintf("/%s/%s", p.PublishedOn.In(config.TimeZone).Format("2006/01/02"), p.Slug())
        },
        "CategoryPath": func(p *post.Post) string {
            return fmt.Sprintf("/category/%s", p.Category)
        },
    }
}
