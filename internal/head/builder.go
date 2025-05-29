// internal/head/builder.go
//
// The Builder collects everything that should appear inside a page’s
// <head> element.  It is scoped to a single request (or render call).  Core
// and modules push tags into the builder, then the theme’s base layout
// decides where to emit each slice.
//
// Features
// --------
//   - SetTitle           – single <title> tag (last call wins).
//   - Meta, Link, Script – arbitrary tags with optional deduplication.
//   - JSONLD             – stores raw JSON-LD strings and wraps them in
//     <script type="application/ld+json">…</script>.
//   - Render helpers     – concat methods that return template.HTML.
package head

import (
	"html/template"
	"strings"
	"sync"
)

// Builder is **not** safe for concurrent writes from multiple goroutines,
// but typical use is one goroutine per request, so a simple mutex is enough.
type Builder struct {
	mu sync.Mutex

	// Single-value fields
	title string

	// Multi-value slices
	metas   []string
	links   []string
	scripts []string
	jsonLD  []string

	// seen tracks keys for deduplication (optional).
	seen map[string]struct{}
}

func New() *Builder {
	return &Builder{seen: make(map[string]struct{})}
}

// ------------------------------------------------------------------
// Single-value helper
// ------------------------------------------------------------------

// SetTitle overrides the page <title>.  The last caller wins.
func (b *Builder) SetTitle(t string) {
	b.mu.Lock()
	b.title = t
	b.mu.Unlock()
}

// Title returns a fully formed <title> tag or an empty string.
func (b *Builder) Title() template.HTML {
	if b.title == "" {
		return ""
	}
	escaped := template.HTMLEscapeString(b.title)
	return template.HTML("<title>" + escaped + "</title>")
}

// ------------------------------------------------------------------
// Slice helpers with deduplication
// ------------------------------------------------------------------

func (b *Builder) Meta(tag string)   { b.add("meta:"+tag, &b.metas, tag) }
func (b *Builder) Link(tag string)   { b.add("link:"+tag, &b.links, tag) }
func (b *Builder) Script(tag string) { b.add("script:"+tag, &b.scripts, tag) }
func (b *Builder) JSONLD(js string)  { b.add("jsonld:"+hash(js), &b.jsonLD, js) }

func (b *Builder) add(key string, tgt *[]string, tag string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, dup := b.seen[key]; dup {
		return
	}
	b.seen[key] = struct{}{}
	*tgt = append(*tgt, tag)
}

// hash creates a short, stable key for JSON-LD strings.
func hash(s string) string {
	if len(s) > 32 {
		return s[:32]
	}
	return s
}

// ------------------------------------------------------------------
// Rendering helpers called from theme templates
// ------------------------------------------------------------------

func (b *Builder) Metas() template.HTML   { return concat(b.metas) }
func (b *Builder) Links() template.HTML   { return concat(b.links) }
func (b *Builder) Scripts() template.HTML { return concat(b.scripts) }

// JSON returns all JSON-LD blocks wrapped in <script> tags.
func (b *Builder) JSON() template.HTML {
	if len(b.jsonLD) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, js := range b.jsonLD {
		sb.WriteString(`<script type="application/ld+json">`)
		sb.WriteString(js)
		sb.WriteString(`</script>`)
	}
	return template.HTML(sb.String())
}

// concat joins pre-escaped tags without a separator.
func concat(sl []string) template.HTML {
	return template.HTML(strings.Join(sl, ""))
}
