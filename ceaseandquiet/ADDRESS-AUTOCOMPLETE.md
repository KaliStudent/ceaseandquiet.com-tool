# Expanded profile and address suggestions

Version 0.4.0 collects first name, last name, email, street address, city, state, ZIP code, date of birth, and phone. Birth date and phone are optional in this preview; verified broker procedures determine which details a real request needs. The other seven fields are required when encrypting the vault. State retains the internal field name `region`.

## Vault compatibility

The expanded profile uses encryption envelope version 3 and a new authenticated context. Versions 0.2.0/0.3.0 used a four-field envelope. Older saved vaults are preserved and rejected by this build rather than silently reinterpreted. Users must explicitly erase and recreate them. No automatic deletion or plaintext migration occurs.

All nine fields participate in encryption, lock/expiry cleanup, and atomic document-pinned form fill. The new extension needs the matching nine-field website; it refuses to fill the older four-field page.

## Address lookup

Both forms offer an enabled-by-default address suggestions box. Users can switch it off. After typing at least four characters, a 400 ms debounce calls the same-origin `/api/v1/address-suggestions` endpoint with a JSON body containing only `text`. Selecting a suggestion fills street address, city, state, and ZIP. Users must check the result and add unit details manually.

The extension performs lookup inside the approved page with its temporary activeTab grant. It retains connect-src none, no permanent host permissions, and no direct provider access. Open the approved preview page before using suggestions in the popup; otherwise manual entry remains available.

The server forwards only the partial address to Geoapify's Address Autocomplete API. Names, email, birth date, and phone are not included. Geoapify also sees the server's connection metadata. Suggestions are not proof of residence, deliverability, or successful broker matching.

Queries and suggestions are not persisted or logged by this application. POST bodies avoid putting user queries in our URLs. The Geoapify upstream API uses query parameters; operational proxies must not log its outgoing URLs. Provider response data is reduced to the five address fields needed by the UI. Response caching is disabled.

## Connect the key

1. Create a Geoapify project and obtain its API key for Address Autocomplete. Do not commit it, put it in frontend JavaScript, or include it in a review package.
2. On the approved staging server, create root-owned mode-0600 `/etc/ceaseandquiet-staging/address.env` with `CQ_GEOAPIFY_API_KEY=<your-key>`. The provided service unit optionally loads this file. Restrict the Geoapify key to the server's verified outgoing IP when possible.
3. Deploy the matching reviewed server build/unit and restart only `ceaseandquiet-staging.service` after approval. Existing cPanel sites need no restart or DNS change.
4. Verify a synthetic lookup through the authenticated staging site and native Chrome popup. Keep real personal details out of test evidence.

Without a key, lookup returns 503 and the UI offers manual entry. Errors, origin rejection, quota limits, and no matches also leave manual entry available. No fake suggestions are shown.

The proxy accepts only JSON POST, one text field of 4–200 bytes, a maximum 1024-byte body, and an approved Origin. It has a four-second provider timeout, rejects redirects, caps response reads at 64 KiB, returns at most five normalized suggestions, and permits at most 60 provider requests per minute across this process. Review provider quotas before enabling it for more users.

This is the first intentional server-side address input in the preview. It is limited to address lookup; there is still no full-profile submission, SMTP integration, or broker delivery endpoint. Deployment and external-provider activation remain separate from local implementation.

## Verification scope

Automated tests cover encryption/validation of the expanded profile, all-field fill checks, enable/disable and selection behavior, stale-result cancellation, provider response normalization, input/origin rejection, disabled-provider fallback, redacted failures, and rate limiting. Browser automation was denied by its URL policy during this change, so visual/native verification of version 0.4.0 remains pending. Earlier native reports cover earlier versions only.

Provider references: https://apidocs.geoapify.com/docs/geocoding/address-autocomplete/ and https://myprojects.geoapify.com/help/api-keys/ .
