// Package views holds the HTML templates and static files. They are compiled into the binary,
// so the image doesn't need a views folder next to it.
package views

import "embed"

// Files contains the templates (*.html, layouts/) and the static files served under /static.
//
//go:embed *.html layouts static
var Files embed.FS
