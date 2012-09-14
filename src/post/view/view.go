package view

import (
    "fmt"
    T "html/template"
    "post"
    "time"
)

func FuncMap() T.FuncMap {
    return T.FuncMap{
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
        "MonthlyPath": func(t time.Time) string {
            return t.Format("/2006/01")
        },
        "TagPath": func(tag string) string {
            return fmt.Sprintf("/tag/%s", tag)
        },
    }
}
