package post

import (
    "bytes"
    "database/sql"
    "encoding/csv"
    "store"
    "time"
)

type Post struct {
    Id                                int
    Title, Category, Description      string
    Renderer, Body, BodyHtml          string
    Published, Announced              bool
    Slugs, Terms, Tags, Images        []string
    PublishedOn, CreatedAt, UpdatedAt time.Time
}

func (p Post) Slug() string {
    return string(p.Slugs[0])
}

type NotFound string

func (n NotFound) Error() string {
    return string(n)
}

func pgVarcharArrayToStringSlice(data []byte) ([]string, error) {
    if len(data) <= 2 {
        return nil, nil
    }
    data = bytes.Trim(data, "{}")
    return csv.NewReader(bytes.NewReader(data)).Read()
}

func scanPost(r *sql.Rows) (*Post, error) {
    post := new(Post)
    var slugs, terms, tags, images []byte
    err := r.Scan(&post.Id,
        &post.Title, &post.Category, &post.Description,
        &post.Renderer, &post.Body, &post.BodyHtml,
        &post.Published, &post.Announced,
        &slugs, &terms, &tags, &images,
        &post.PublishedOn, &post.CreatedAt, &post.UpdatedAt)
    post.Slugs, err = pgVarcharArrayToStringSlice(slugs)
    if err != nil {
        return nil, err
    }
    post.Terms, err = pgVarcharArrayToStringSlice(terms)
    if err != nil {
        return nil, err
    }
    post.Tags, err = pgVarcharArrayToStringSlice(tags)
    if err != nil {
        return nil, err
    }
    post.Images, err = pgVarcharArrayToStringSlice(images)
    if err != nil {
        return nil, err
    }
    return post, err
}

func scanPosts(rows *sql.Rows) ([]*Post, error) {
    posts := make([]*Post, 0, 6)
    for rows.Next() {
        post, err := scanPost(rows)
        if err != nil {
            return nil, err
        }
        posts = append(posts, post)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return posts, nil
}

func FindByTag(tag string) ([]*Post, error) {
    db, err := store.Connect()
    if err != nil {
        return nil, err
    }
    defer store.Disconnect()
    rows, err := db.Query("SELECT * FROM published_posts WHERE $1 = ANY(tags)", tag)
    if err != nil {
        return nil, err
    }
    return scanPosts(rows)
}
