package page

import (
    "fmt"
    "store"
    "time"
)

type Page struct {
    Id                   int
    Title, Slug, Body    string
    CreatedAt, UpdatedAt time.Time
}

type NotFound string

func (n NotFound) Error() string {
    return string(n)
}

func FindBySlug(slug string) (*Page, error) {
    db, err := store.Connect()
    if err != nil {
        return nil, err
    }
    defer store.Disconnect()
    row := db.QueryRow("SELECT * FROM pages WHERE pages.slug = $1 LIMIT 1", slug)
    page := new(Page)
    if err := row.Scan(&page.Id, &page.Title, &page.Slug, &page.Body, &page.CreatedAt, &page.UpdatedAt); err != nil {
        if err == store.NoRows {
            err = NotFound(fmt.Sprintf("Page %#v could not be found", slug))
        }
        return nil, err
    }
    return page, nil
}
