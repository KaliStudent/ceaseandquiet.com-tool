# Staging vault acceptance

Version 0.4.0 authorizes only the root page at `https://abcw.xyz:8443/` and the existing localhost preview. The scheme, hostname, port, path, and query restrictions are checked before injection and within the serialized injected function. The extension still uses activeTab rather than permanent host access and cannot submit the form.

This build has automated coverage; its native staging acceptance is pending. The earlier NATIVE-CHROME-RESULTS.md covers version 0.2.0 on localhost and must not be read as evidence for this new staging build.

## Owner setup

1. Open the staging address in Chrome and sign in using the separately supplied staging credentials. Do not include credentials in the URL or screenshots.
2. Load this package's extension folder as an unpacked extension in a development profile. If updating the previously installed extension, choose its new folder or copy the new extension files into the old installation folder and reload it. Do not run both versions during the test.
3. Confirm the displayed extension version is 0.4.0. Return to the staging page and open the extension using its toolbar action.

## Native checks with synthetic details

- [ ] Create a synthetic profile and unique test passphrase; save locks and clears the popup.
- [ ] Reopen: saved vault remains locked. A wrong passphrase fails without revealing profile values.
- [ ] Unlock and verify values. Explicit Lock now clears them.
- [ ] Fill the empty staging form. All nine values appear, the popup locks, and the default erase-after-fill removes the saved vault.
- [ ] Confirm no page submission occurred; values exist only in the browser form. Clear the form.
- [ ] Repeat with erase-after-fill unchecked. The encrypted vault persists and can be unlocked again.
- [ ] Attempt another fill while any page field is occupied; reject without overwriting or erasing the saved vault.
- [ ] Attempt fill on the ordinary port-443 site or another local path/port; reject.
- [ ] Leave an unlocked popup untouched for two minutes; it locks and clears revealed values.
- [ ] Confirm page values clear on hiding/leaving and after the page's two-minute timer.
- [ ] Erase the synthetic vault and clear the page after testing.

Record the package source commit, Chrome version, and observed results. Do not record passphrases, access credentials, or real personal details. Native navigation-race and immediate-close-during-unlock checks remain part of BETA-CHECKLIST.md.

## Development status

The live directory and read-only API are deployed separately. This extension package does not deploy backend changes. Removal requests, account authentication, server-side personal-data intake, and broker email delivery are not enabled by this change.

## Version 0.4.0 additions

- [ ] Recreate older vaults explicitly; this version uses a new nine-field envelope.
- [ ] Verify street address, city, state, ZIP, birth date, and phone survive encryption and fill correctly.
- [ ] Confirm suggestions stay off until selected, and manual entry works without an API key.
- [ ] With the provider connected, select a synthetic-address suggestion in the site and popup; check street/city/state/ZIP.
- [ ] Confirm no lookup includes name, email, birth date, phone, or the vault passphrase.
- [ ] Confirm clearing/locking/hiding removes pending suggestions and no late result repopulates them.
