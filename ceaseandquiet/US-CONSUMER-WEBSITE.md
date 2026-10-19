# US consumer website rebuild

## Product structure
- / is the service explanation followed directly by recent privacy reporting.
- /privacy-scan/ contains the US-resident profile form, address suggestions, Turnstile, email confirmation, optional encrypted progress, and broker/advertising education.
- /consumer-protection/ explains credit reports versus scores, links to Experian's score service and AnnualCreditReport.com, explains breach checks through Have I Been Pwned, and links to FTC identity recovery guidance.

The public navigation has exactly these three destinations. The homepage no longer presents the broker directory or vault as the product. The approved dark palette and transparent, white-text buttons with bright gradient hover borders are preserved.

## Authorization and US scope
Submitting the single scan/removal action authorizes both supported-broker searches and removal/opt-out requests on the resident's behalf. Residency/own-information confirmation, a US state or territory, a US ZIP format, the current combined-consent version, a real bot session, and server-confirmed email ownership are required before any configured workflow can start. The server enforces these independently of browser fields. There is no scan-only authorization mode.

Address suggestions request US results and reject non-US provider records; US territories can also be entered manually. Saved profiles are optional. Unsaved form details remain in the tab when the visitor checks their email, and clear on navigation/reset or after 15 minutes of inactivity. Passphrases clear when the tab is hidden. Revealed encrypted profiles retain their two-minute/hide lock.

## Privacy news
The server fetches the fixed EFF RSS source, limits response size and time, filters privacy-related stories, validates original-publisher HTTPS links and dates, and returns at most six headlines sorted by publication date. It caches results for 15 minutes and backs off on upstream failure. Older cached results are marked stale. A missing feed is reported honestly and links to the publisher remain available. No publisher scripts, tracking pixels, or third-party news embeds are loaded.

## Consumer and affiliate resources
Official resources are linked directly. Credit reports are distinguished from credit scores. Have I Been Pwned is an external check; no visitor email is appended to or transmitted through those links. The affiliate section stays hidden until the owner supplies exact provider links in site/affiliate-links.json. Links preserve supplied URLs, require HTTPS, use sponsored/noopener/noreferrer, and are preceded by a commission disclosure. No affiliate providers or referral codes were invented.

## Service readiness
This rebuild does not invent scan matches from the broker directory. No production presence-scan/removal adapter exists yet. The status endpoint therefore reports unavailable, the submit control is disabled, and a direct request fails closed without accepting a personal profile into a fake queue. The authorization boundary and future adapter contract are tested with synthetic data. No real broker email is sent. Email confirmation still awaits the noreply sender mailbox setup review. Those are remaining implementation milestones, not completed features.

## Evidence
36 Node tests passed. Targeted Go tests and vet passed for news, scan authorization, address suggestions, preview/staging, bot/email verification, and resource limits. Browser automation was previously denied by URL policy, so no visual browser validation is claimed. Live deployment checks will verify all three routes, protected assets, privacy news, authentication, and unavailable scan status. Final security review remains pending.

Source references: consumer.ftc.gov people-search/online-tracking/credit guidance; consumerfinance.gov credit-report guidance; Experian credit-score page; haveibeenpwned.com; IdentityTheft.gov; EFF original RSS headlines. User-facing page links identify the sources near the relevant guidance.