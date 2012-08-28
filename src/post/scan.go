package post

import (
    "bytes"
    "database/sql"
    "encoding/csv"
    "fmt"
)

type columns interface {
    Columns() ([]string, error)
}

type scanner interface {
    Scan(...interface{}) error
}

func pgVarcharArrayToStringSlice(data []byte) ([]string, error) {
    if len(data) <= 2 {
        return nil, nil
    }
    data = bytes.Trim(data, "{}")
    return csv.NewReader(bytes.NewReader(data)).Read()
}

func scanPost(s scanner) (*Post, error) {
    post := new(Post)
    var slugs, terms, tags, images []byte
    var rank float32
    targets := []interface{}{&post.Id,
        &post.Title, &post.Category, &post.Description,
        &post.Renderer, &post.Body, &post.BodyHtml, &post.BodyTruncated,
        &post.Published, &post.Announced,
        &slugs, &terms, &tags, &images,
        &post.PublishedOn, &post.CreatedAt, &post.UpdatedAt}
    if c, ok := s.(columns); ok {
        cols, err := c.Columns()
        if err != nil {
            return nil, err
        }
        switch n := len(cols); n {
        case 17:
            // Do nothing, but validate that it's fine
        case 18:
            targets = append(targets, &rank)
        default:
            return nil, fmt.Errorf("Wrong number of columns %d", n)
        }
    }
    err := s.Scan(targets...)
    if err != nil {
        return nil, err
    }
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
