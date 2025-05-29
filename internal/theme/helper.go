package theme

import "html/template"

// FuncMap provides template helpers scoped to themes.
// For now we expose only the `asset` helper so there is no
// import cycle between theme and tenant packages.
func FuncMap(assetFn func(string) string) template.FuncMap {
	return template.FuncMap{
		"asset": assetFn,
	}
}
