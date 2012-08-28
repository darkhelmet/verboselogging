package view

import (
    "fmt"
    T "html/template"
    "post"
)

func FuncMap() T.FuncMap {
    return T.FuncMap{
        "CategoryPath": func(p *post.Post) string {
            return fmt.Sprintf("/category/%s", p.Category)
        },
        "TagPath": func(tag string) string {
            return fmt.Sprintf("/tag/%s", tag)
        },
    }
}
