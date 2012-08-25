package config

import (
    "fmt"
    "github.com/darkhelmet/env"
    "log"
)

var (
    user            = env.StringDefault("USER", "")
    DatabaseUrl     = env.StringDefault("DATABASE_URL", fmt.Sprintf("postgres://localhost/darkblog2_development?sslmode=disable"))
    Port            = env.IntDefault("PORT", 5000)
    CanonicalHost   = env.StringDefaultF("CANONICAL_HOST", func() string { return fmt.Sprintf("localhost:%d", Port) })
    AssetHost       = env.StringDefaultF("ASSET_HOST", func() string { return fmt.Sprintf("http://%s", CanonicalHost) })
    LogFlags        = env.IntDefault("LOG_FLAGS", log.LstdFlags|log.Lmicroseconds)
    SiteTitle       = "Verbose Logging"
    SiteDescription = "software development with some really amazing hair"
    SiteContact     = "darkhelmet@darkhelmetlive.com"
)
