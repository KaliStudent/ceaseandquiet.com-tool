# Cease & Quiet local beta candidate

This candidate provides a local broker directory and an encrypted Chrome vault for a sample form. It is staged for the owner's final review.

## Run from source

From the repository root, run go run ./cmd/cq-preview and open http://127.0.0.1:8787/. The Go process binds only to loopback, serves embedded allowlisted assets, rejects unexpected Host headers, and prevents form submission through browser CSP. No profile data is accepted.

GET /api/v1/brokers supports q, region (us/eu/global), limit (1–100), and offset. Responses provide public metadata, total count, and paging state. GET /api/v1/brokers/{id} returns one record. Contact channels remain unreviewed and are omitted.

## Load the extension

Use a clean development Chrome profile. Load ceaseandquiet/extension as an unpacked MV3 extension. Use sample details and a unique test passphrase. The candidate fills only the empty marked form at http://127.0.0.1:8787/. It has no production origin access and never submits the form.

Version 0.2.0 uses a new authenticated envelope format. Older prototype vaults must be erased and recreated. The vault locks after save, after successful fill, on close/hide, after a failed unlock, or after two minutes. Erase saved vault after successful fill is selected by default. Clearing the popup or deleting saved ciphertext cannot recall values already transferred into a page.

## Review and rollout

See SECURITY-REVIEW.md for data flow, safeguards, evidence, limitations, and review focus. Complete BETA-CHECKLIST.md in actual Chrome. Follow ROLLOUT.md for the owner's final review and beta release.

Core native Chrome vault and fill checks passed in an existing Windows Chrome profile; see NATIVE-CHROME-RESULTS.md. The complete clean-profile checklist and final security review remain pending. Module tests also exercise failure paths using simulated Chrome APIs. This is a beta candidate, not a production security claim.

## Package on Windows

Run ceaseandquiet/scripts/package-beta.ps1 -OutputDirectory <absolute output directory>. It builds a trimmed local preview executable, copies only the required extension files and review documents, records the source commit and working-tree state, then creates a ZIP and checksum. It refuses to package a dirty checkout.
