// fs.go holds tiny helpers for walking the filesystem when template glob
// patterns such as “**/*.html” are not available in the Go standard library.
// The key export is CollectHTML, which returns a slice of absolute paths for
// every .html file under the supplied directory.
package theme

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// CollectHTML walks rootDir recursively and returns a list of *.html paths.
// Paths are returned in slash form (even on Windows) so they can be fed
// straight into template.ParseFiles or template.ParseGlob.
//
// Callers typically pass:
//
//	files, _ := CollectHTML("/themes/ocean/templates")
//	tpl.ParseFiles(files...)
func CollectHTML(rootDir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil { // propagate filesystem errors immediately
			return err
		}
		// Skip directories quickly.
		if d.IsDir() {
			return nil
		}
		// We care only about *.html files.
		if strings.HasSuffix(strings.ToLower(d.Name()), ".html") {
			files = append(files, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
