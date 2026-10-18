# Website form security

## Website workflow
The website supports profile entry, address suggestions, an encrypted local profile, unlock/edit/save, lock, and erase without installing an extension. Only ciphertext is saved in this website origin's localStorage. No full profile or passphrase is sent to the server. Profile reveal ends after two minutes, hide/navigation, reset, or explicit lock. Clearing site storage removes the saved envelope. The owner's final security review is pending.

## Bot protection
Cloudflare Turnstile runs with action profile_form and hostname abcw.xyz. The secret stays in the protected service environment file. Only the public site key is returned by the authenticated config endpoint. Tokens are verified server-side with a four-second timeout, no redirects, a bounded response, and exact success/hostname/action checks. Cloudflare enforces token expiry and single use. A verified challenge creates a random, HttpOnly, Secure, SameSite=Strict API cookie lasting 30 minutes. Process restarts invalidate sessions.

When Turnstile is configured, address lookups require that session. Email start/confirm/status also require it. Unknown inputs and a filled honeypot are rejected. The website does not send a complete profile to any endpoint. A future broker-delivery endpoint MUST call RequireVerifiedEmail with the actual request email before delivery; no broker-delivery route exists in this build.

The staging guard limits each TCP peer to 180 requests/minute, all peers to 600/minute, concurrent handlers to 32, and bodies to 16 KiB. Address bodies are limited to 1024 bytes and 60 provider requests/minute. Forwarded IP headers are ignored. Rate buckets use process-salted IP hashes retained for the current minute. Session and email-quota maps are capped at 2048 entries and pruned on requests. These application bounds do not replace network-level DDoS protection.

## Email confirmation
The server generates a random six-digit code, binds its hash to the human session and exact email address, expires it after ten minutes, and allows at most five attempts. A successful code is consumed. Changing the email requires confirmation again. Only hashes are kept in temporary memory; codes and addresses are not logged or persisted. Local-part case remains significant for ownership binding. Case variants share the resend quota.

Resends require a one-minute cooldown. A destination can receive at most three codes/hour; the process can send at most 20/minute. Failed mail sends also use the quota. Sender transport uses authenticated TLS SMTP and a ten-second deadline. There is no background retry loop. Verification email uses noreply@ceaseandquiet.com after its mailbox is configured. Until then, email controls clearly show that mail is not connected and sending fails closed. No confirmation email or broker email has been sent by this build.

## Protected configuration
The unit optionally loads /etc/ceaseandquiet-staging/form-security.env for CQ_TURNSTILE_SITE_KEY and CQ_TURNSTILE_SECRET. The optional verification-mail.env file accepts CQ_VERIFICATION_FROM, CQ_VERIFICATION_SMTP_HOST, CQ_VERIFICATION_SMTP_PORT, CQ_VERIFICATION_SMTP_USER, CQ_VERIFICATION_SMTP_PASSWORD, and CQ_VERIFICATION_EMAIL_ENABLED=true. Both files must be root-owned mode 0600. Do not commit their contents, print them, or include them in reports. Missing mail configuration does not disable bot protection.

## Evidence and remaining work
31 Node tests exercise vault/form behavior, delayed-operation cancellation, website-only profile storage, website bot/session interaction, email-change status clearing, address transport policy, and the date calendar trigger. Go tests cover verification binding, single use, expiry, attempt limits, resend cooldown, unavailable sender, malformed inputs, cross-origin refusal, protected address access, and application resource limits. The local Windows runtime lacks the CGO configuration needed for Go's race detector; no local race-test claim is made. GitHub CI runs separately.

The live website browser check requires user confirmation because browser automation was denied by URL policy. Native calendar styling is limited; a custom calendar matching the full palette remains a requested design refinement. Email activation awaits review and setup of the dedicated noreply sender mailbox. Broker request review/delivery and final security review remain subsequent milestones.

Provider documentation: https://developers.cloudflare.com/turnstile/get-started/server-side-validation/ and https://developers.cloudflare.com/turnstile/reference/content-security-policy/ .