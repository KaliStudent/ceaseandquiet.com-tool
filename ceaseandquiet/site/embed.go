// Package previewsite contains only the public preview assets.
package previewsite

import "embed"

//go:embed index.html privacy-scan.html consumer-protection.html styles.css site.js home.js scan.js consumer.js affiliate-links.json scan-workflow.mjs form-tools.mjs profile-vault.mjs form-security.mjs
var Files embed.FS
