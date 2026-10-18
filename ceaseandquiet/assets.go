// Package cqassets shares the address UI module with the embedded website.
package cqassets

import "embed"

//go:embed extension/address.mjs extension/vault.mjs
var Files embed.FS
