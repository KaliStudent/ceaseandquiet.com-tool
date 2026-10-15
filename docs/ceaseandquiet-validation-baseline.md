# CeaseAndQuiet — Eraser validation baseline

Status: **development only; no live requests authorized**.

## Source and attribution

This branch derives from `drumandbytes/eraser`. Preserve the repository's MIT license and the original author/maintainer notices. Do not publish user profiles, mailbox credentials, authorization documents, or removal histories.

## Observed local baseline (provided by project owner)

Commands executed on the owner's Windows computer:

- `eraser.exe validate-brokers`: main embedded list structurally valid (**723**); embedded verified list structurally valid (**60**).
- `eraser.exe audit-brokers`: **552** alive, **48** email-dead, **38** website-dead, **85** unknown, **0** skipped. Total: 723.
- `eraser.exe list-brokers`: printed broker directory records.

These results reflect the owner's local tool and network environment, not an independent CI run. An audit status of "alive" does **not** demonstrate that a privacy-removal channel works or that any personal record has been deleted. DNS/network failures need independent review.

## Validation gates for an eventual public service

1. **Structural integrity:** all required fields, duplicate IDs, region/category values, and URL/email syntax. Keep the upstream validator and tests.
2. **Contact-channel review:** independently review official broker privacy channels, response handling, redirects, subsidiaries, and opt-out procedures. Track verification date, method, source URL, and reviewer. Never automatically purge entries solely because one network audit failed.
3. **Template safety:** ensure legal claims are appropriate to the user's actual jurisdiction and request. Show exact destination, disclosed fields, and full message before explicit approval.
4. **Delivery isolation:** tests use synthetic profiles and controlled inboxes only. Never send a test email to a real broker. Keep outbound email **disabled by default** in CeaseAndQuiet.
5. **Outcome accuracy:** separately record drafted, approved, submitted, delivered, bounced, needs action, broker-confirmed, and independently rechecked. Do not equate "sent" with "deleted".
6. **Multi-user safety:** require authentication, strong authorization and tenant isolation, encrypted sensitive data, limited retention, and a documented deletion process before any public personal-data intake.
7. **Abuse controls:** protect against unauthorized third-party removal requests, bulk unsolicited delivery, account enumeration, and SSRF through broker-configured URLs.

## Initial test commands (no sending)

```powershell
go test ./internal/broker/...
go run ./cmd/eraser validate-brokers
```

Use Go **1.26** as specified by `go.mod`. These checks must complete before enabling any integration tests. Do **not** invoke `eraser send` without `--dry-run`; the automated workflow intentionally excludes even dry-run email generation until a synthetic test configuration is reviewed.

## Next milestone

Build a broker contact-review report with dated verification evidence and a fail-closed request authorization gate, then test request formatting against owned mailboxes. Keep live delivery entirely separate from CI.
