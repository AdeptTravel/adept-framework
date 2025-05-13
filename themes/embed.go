package themes

import "embed"

// embed every template file under themes/*
//
//go:embed minimal/views/*.html
var FS embed.FS
