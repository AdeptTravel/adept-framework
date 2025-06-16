// internal/routing/slug.go
//
// Slug and path helpers.
//
// • MakeSlug(title) ─ converts arbitrary text into a URL-safe slug restricted
//   to ASCII a-z, 0-9 and “-” (English-only requirement).
// • BuildPath(parent, slug) ─ joins parent path + slug with a single “/” and
//   guarantees exactly one leading slash.
//
// Rules (MakeSlug)
// ----------------
// 1. Lower-case everything.
// 2. Convert any run of non-[a-z0-9] characters to one “-”.  That strips
//    spaces, punctuation, emoji, and non-ASCII.
// 3. Collapse consecutive “-” to a single “-”.
// 4. Trim leading / trailing “-”.
// 5. If the result is empty, return "item".
//
// Notes
// -----
// • No Unicode transliteration because site is English-only for now.
// • Slugs are max 100 runes; callers may truncate earlier if they prefer.

package routing

import (
	"strings"
)

// MakeSlug converts title → lower-kebab ASCII.
func MakeSlug(title string) string {
	var b strings.Builder
	b.Grow(len(title))

	lastWasDash := false
	for _, r := range strings.ToLower(title) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastWasDash = false
		default:
			// any non-ASCII or punctuation becomes a single dash
			if !lastWasDash {
				b.WriteRune('-')
				lastWasDash = true
			}
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "item"
	}
	if len(slug) > 100 {
		slug = slug[:100]
		// trim trailing dash if the cut landed on one
		slug = strings.TrimRightFunc(slug, func(r rune) bool { return r == '-' })
	}
	return slug
}

// BuildPath joins parent + slug ensuring exactly one leading slash and no
// duplicate separators.
func BuildPath(parent, slug string) string {
	parent = strings.Trim(parent, "/")
	slug = strings.Trim(slug, "/")

	switch {
	case parent == "" && slug == "":
		return "/"
	case parent == "":
		return "/" + slug
	case slug == "":
		return "/" + parent
	default:
		return "/" + parent + "/" + slug
	}
}
