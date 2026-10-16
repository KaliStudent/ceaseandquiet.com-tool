// Package previewsite contains only the public preview assets.
package previewsite

import "embed"

//go:embed index.html styles.css site.js
var Files embed.FS
