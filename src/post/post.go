package post

import (
    "config"
    "fmt"
    "log"
    "os"
    "time"
)

var (
    logger = log.New(os.Stdout, "[post] ", config.LogFlags)
)

type Post struct {
    Id                                      int
    Title, Category, Description            string
    Renderer, Body, BodyHtml, BodyTruncated string
    Published, Announced                    bool
    Slugs, Terms, Tags, Images              []string
    PublishedOn, CreatedAt, UpdatedAt       time.Time
}

func (p *Post) Author() string {
    return config.SiteAuthor
}

func (p *Post) FeedAuthor() string {
    return fmt.Sprintf("%s (%s)", config.SiteContact, config.SiteAuthor)
}

func (p *Post) Slug() string {
    return p.Slugs[0]
}

func (p *Post) Canonical() string {
    return fmt.Sprintf("/%s/%s", p.PublishedOn.In(config.TimeZone).Format("2006/01/02"), p.Slug())
}

func (p *Post) Related() []*Post {
    posts, err := queryMany(`
        SELECT published_posts.*
        FROM
            published_posts,
            (SELECT terms AS oterms FROM published_posts WHERE id = $1) AS post
        WHERE terms && oterms
        AND id <> $1
        ORDER BY ARRAY_LENGTH(array_intersect(terms, oterms), 1) DESC
        LIMIT 6`, p.Id)
    if err != nil {
        logger.Printf("failed retrieving related posts: %s", err)
        return nil
    }
    return posts
}
