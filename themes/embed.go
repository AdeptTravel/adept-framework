// themes/embed.go
package themes

import "embed"

//go:embed minimal/views/*.html
var FS embed.FS
