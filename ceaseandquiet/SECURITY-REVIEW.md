# Cease & Quiet local beta review packet

Status: **candidate for final review; synthetic details only**. Final review will be performed using the owner's chosen security tools/team. This document does not claim an independent audit.

## What the candidate does

The loopback preview serves embedded static assets and public broker metadata. The MV3 popup encrypts four profile fields on device, unlocks for at most two minutes, fills the exact local sample form after an explicit click, then locks. The user can choose to delete saved ciphertext after successful fill (selected by default).

There is no server profile intake, form submission, SMTP, scan, background extension worker, telemetry, remote script, sync storage, or external messaging in this flow. The upstream Eraser CLI retains its existing send functionality; this separate preview does not invoke or import its sender.

## Data and trust boundaries

1. The user enters a synthetic profile and passphrase into the extension popup.
2. PBKDF2-SHA256 with 600,000 iterations derives an AES-256-GCM key. A random 16-byte salt and 12-byte nonce are generated per save. The cipher authenticates a fixed application context.
3. Only a versioned encryption envelope is persisted in Chrome local extension storage. Storage access is limited to trusted extension contexts.
4. Decrypted profile strings live in the popup during an unlocked session. Closing, hiding, locking, a failed unlock, expiry, and successful fill clear revealed fields and discard session references.
5. Fill first probes the empty top-level local form without profile arguments, then targets that exact Chrome document ID. The injected function checks origin, path, query, form marker, and all target fields again before writing.
6. Once filled, profile values exist in the page DOM. They clear on page hide/leave, explicit reset, or a two-minute timer.

No JavaScript implementation can promise forensic erasure of immutable strings, browser history, memory copies, device backups, or a profile already exposed to a trusted page. Deleting ciphertext is not secure disk erasure.

## Findings addressed during development

- Added submission prevention in the page and a CSP with form-action none, including when JavaScript fails.
- Cleared stale revealed fields and passphrases on failed unlock or locking.
- Added an absolute session deadline and cancellation of late unlock results.
- Removed the unreviewed production fill origin. This candidate accepts only http://127.0.0.1:8787/ (optional anchor).
- Pinned fill to the document that passed preflight and rechecked the destination within the injected function.
- Rejected malformed vaults, unsupported metadata, weak/changed KDF settings, oversized values, and unexpected profile fields.
- Restricted local storage to trusted extension contexts and blocked extension network connections with CSP.
- Replaced static directory serving with embedded allowlisted assets; rejected unexpected Host headers and non-loopback listeners.
- Added visible busy/locked states and failure handling for ciphertext deletion after a successful fill.

## Automated coverage

Run from the repository root:

- go test ./internal/readonlyapi ./internal/preview ./cmd/cq-preview
- go vet ./internal/readonlyapi ./internal/preview ./cmd/cq-preview
- node --test ceaseandquiet/tests/*.test.mjs
- go run ./cmd/eraser validate-brokers

The Node tests exercise actual WebCrypto and popup/session logic, with simulated Chrome storage and scripting APIs. They cover round trips, randomness, wrong passwords, tampering, metadata rejection, expiry, cancellation, save/reopen/unlock/fill/erase, storage failure, destination rejection, and atomic field validation. They do not establish native Chrome permission behavior.

Go tests cover pagination/filtering, read-only methods, no contact fields, loopback-only binding, Host rejection, private-file exclusion, security headers, and static query rejection.

## Browser evidence and remaining gate

The in-app browser was used to verify the local website, search, pagination, record details, empty results, clear form, and a 390-pixel mobile viewport without horizontal overflow.

**Actual unpacked extension installation, popup behavior, activeTab permission grant, and document-ID injection in Chrome remain to be verified.** No Chrome browser connection is available to this chat. Complete BETA-CHECKLIST.md in a clean Chrome development profile before claiming the extension is beta-ready.

## Review focus for the owner's final tools/team

Review the same commit independently, then reconcile findings:

- Cryptography and envelope validation in extension/vault.mjs.
- Race conditions and cleanup in extension/popup.js.
- Destination authorization and data transfer in extension/fill.mjs.
- Minimum permissions, CSP, and local storage access in extension/manifest.json.
- Browser CSP and local-only serving in internal/preview.
- Accurate privacy claims in UI copy and documentation.
- Whether any logs, errors, artifacts, or directory output expose personal details.

Record a reproducible path for every finding and severity. Fixes produce a new commit and rerun the relevant checks; reviewers should recheck that commit. A passing automated scan is one source of evidence, not a guarantee that a privacy design is correct.
