package view

import "html/template"

// FuncMap holds helpers available in every template.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		// add helpers as you need them
	}
}
