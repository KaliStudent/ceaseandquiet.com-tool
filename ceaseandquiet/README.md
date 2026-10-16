# Cease & Quiet staged prototype

This folder is isolated from the upstream Eraser docs site and command line send flow. It is not a production service.

## Local preview

Run `go run ./cmd/cq-preview` from the repository root and open `http://127.0.0.1:8787`. The Go process serves the static site and two **GET-only** broker metadata endpoints:

- `/api/v1/brokers?limit=25&q=name`
- `/api/v1/brokers/{id}`

The API reads the embedded broker list, not local profiles or a user's edited broker file. It omits contact addresses and opt-out links until independently reviewed. It has no sender, mail configuration, profile, or history code. The preview binds to loopback by default; do not expose it publicly without separate authentication, rate limiting, deployment review, and privacy review.

## Chrome extension

Load `ceaseandquiet/extension` as an unpacked Manifest V3 extension in a development Chrome profile. The popup encrypts a limited profile with PBKDF2-SHA256 (600,000 iterations) and AES-256-GCM before saving ciphertext in `chrome.storage.local`. It requests only storage, activeTab, and scripting permissions. A user must unlock and click Fill on the preview page. It fills the form and never submits it. Erase deletes the extension's saved ciphertext; copies already placed in page fields, browser sync/backups, and process memory are outside that deletion claim.

The extension is an early prototype. It has not had a security audit and should use synthetic test details until that review is complete. Public intake, broker email delivery, scans, account management, and DNS/VPS work are outside this branch.
