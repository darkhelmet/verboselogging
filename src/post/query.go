package post

import (
    "config"
    "fmt"
    "store"
    "strconv"
    "strings"
    "time"
)

func stringsToInts(inputs ...string) ([]int, error) {
    ints := make([]int, len(inputs))
    var err error
    for i, str := range inputs {
        ints[i], err = strconv.Atoi(str)
        if err != nil {
            return nil, err
        }
    }
    return ints, nil
}

func queryMany(query string, args ...interface{}) ([]*Post, error) {
    db, err := store.Connect()
    if err != nil {
        return nil, err
    }
    defer store.Disconnect()

    rows, err := db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    return scanPosts(rows)
}

func FindByPermalink(year, month, day, slug string) (*Post, error) {
    parsed, err := stringsToInts(year, month, day)
    if err != nil {
        return nil, err
    }

    db, err := store.Connect()
    if err != nil {
        return nil, err
    }
    defer store.Disconnect()

    startOfDay := time.Date(parsed[0], time.Month(parsed[1]), parsed[2], 0, 0, 0, 0, config.TimeZone).UTC()
    endOfDay := time.Date(parsed[0], time.Month(parsed[1]), parsed[2], 23, 59, 59, 0, config.TimeZone).UTC()
    row := db.QueryRow(`
        SELECT *
        FROM published_posts
        WHERE $1 = ANY(slugs)
        AND published_on
        BETWEEN $2 AND $3 LIMIT 1`, slug, startOfDay, endOfDay)
    post, err := scanPost(row)
    if err != nil {
        if err == store.NoRows {
            err = NotFound(fmt.Sprintf("Post %#v could not be found", slug))
        }
        return nil, err
    }
    return post, nil
}

func FindByTag(tag string) ([]*Post, error) {
    return queryMany("SELECT * FROM published_posts WHERE $1 = ANY(tags)", tag)
}

func FindByCategory(category string) ([]*Post, error) {
    return queryMany("SELECT * FROM published_posts WHERE category = $1", category)
}

func FindLatest(limit int) ([]*Post, error) {
    return queryMany("SELECT * FROM published_posts LIMIT $1", limit)
}

func Search(query string) ([]*Post, error) {
    query = strings.TrimSpace(query)
    if query == "" {
        return nil, nil
    }
    return queryMany(`
        SELECT published_posts.*, (ts_rank((setweight(to_tsvector('simple', title), 'A') || setweight(to_tsvector('simple', description), 'B') || setweight(to_tsvector('simple', ARRAY_TO_STRING(tags, ',')), 'B') || setweight(to_tsvector('simple', body), 'C')), (to_tsquery('simple', ''' ' || $1 || ' ''')), 0)) AS pg_search_rank
        FROM published_posts
        WHERE (((setweight(to_tsvector('simple', title), 'A') || setweight(to_tsvector('simple', description), 'B') || setweight(to_tsvector('simple', ARRAY_TO_STRING(tags, ',')), 'B') || setweight(to_tsvector('simple', body), 'C')) @@ (to_tsquery('simple', ''' ' || $1 || ' '''))))
        ORDER BY pg_search_rank DESC`, query)
}
