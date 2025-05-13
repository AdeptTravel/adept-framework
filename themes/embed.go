package themes

import "embed"

// Embed everything under each theme directory (minimal, future themes).
//
//go:embed minimal/**/*
var FS embed.FS
