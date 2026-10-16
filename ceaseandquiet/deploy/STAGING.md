# Restricted HTTPS staging proposal

Target: `wild.feralchat.tech`, staging origin `https://abcw.xyz:8443`.

## Verified read-only inspection, 2026-10-16

- SSH using the existing trusted IP host key confirmed the server hostname and x86_64 architecture.
- Local DNS and queries to two public resolvers returned the same address for the VPS hostname, abcw.xyz, and ceaseandquiet.com. This is a point-in-time result, not proof of universal propagation.
- cPanel Apache is active; Podman and Docker are installed. No running root-owned Podman containers were listed. This does not establish that other users have no containers.
- TCP ports 8443 and 8787 were not listening. The host INPUT policy was ACCEPT; Hostinger cloud firewall state was not established because its inventory call did not return.
- abcw.xyz belongs to the cPanel account `blueberryrage`. Its current certificate is self-signed; an ordinary HTTPS client rejected the certificate chain.
- systemd version 252 is available. No files were found at the proposed service, deployment, or configuration paths.

No server files, services, firewall rules, or DNS records were changed during this inspection.

## Exact proposed write scope, subject to owner review

1. Request a trusted certificate through the existing cPanel AutoSSL mechanism for account `blueberryrage`. The per-user AutoSSL command can examine the account's other domains and renew their eligible certificates; inspect its provider/account scope before execution. Verify abcw.xyz's chain and hostname afterward. Do not start public staging with the self-signed certificate or ask testers to bypass warnings.
2. Install the new Linux amd64 `cq-staging` executable at `/opt/ceaseandquiet-staging/cq-staging`, root-owned and not writable by its service identity.
3. Generate a unique random access password with at least 32 bytes of entropy. Store only its SHA256 hex digest in root-owned mode-0600 `/etc/ceaseandquiet-staging/auth.sha256`. Deliver the password privately, outside the repository, logs, PR, and this report. This digest design requires a generated high-entropy password, not a memorable user-chosen password.
4. Install the included dedicated systemd unit at `/etc/systemd/system/ceaseandquiet-staging.service`. Validate the unit, reload systemd's unit definitions, and enable/start only this new service. Do not restart Apache, cPanel, Podman, or the VPS.
5. The service listens on TCP 8443, serves TLS directly, validates the exact external Host, and authenticates every website/API request. It wraps the current embedded read-only preview in memory, so it requires no proxy and opens no loopback upstream listener. It runs with a temporary unprivileged identity, read-only filesystem protections, empty capabilities, and resource limits.
6. Confirm external reachability after startup. If an upstream firewall blocks TCP 8443, inspect the actual rule set and obtain approval for a specifically scoped change before opening it. No firewall write is preselected by this proposal.

The public port accepts connection attempts from the internet if upstream firewalls permit them. A nonstandard port is not a security control; TLS and the staging access gate protect content. The access gate is shared staging access, not application user authentication.

## Verification and remaining work

Local Go tests use a real ephemeral TLS server to verify unauthorized denial on assets/API, incorrect credentials, exact Host rejection, authenticated page/API success, HEAD behavior, write-method rejection, private-file exclusion, static-query rejection, privacy headers, and plaintext rejection. The Linux process, systemd unit, real certificate, and external network path still require target-host verification.

After deployment, verify trusted TLS without disabling certificate checks; 401 responses without credentials; page/assets/API with credentials; rejected POST/private-file paths; service identity and limits; and existing website availability. Do not capture credentials or private data in evidence. The service has no access log and no sender dependency.

The existing extension still accepts localhost only. A separate staging build must authorize the exact HTTPS origin and pass its tests before claiming the remote vault flow works. Remaining frontend/backend development proceeds using synthetic data with real broker delivery disabled.

Certificate renewal requires restarting only the staging service to refresh its systemd credential copy. Establish that renewal action before leaving staging running long term.

## Rollback

Disable and stop only `ceaseandquiet-staging.service`. Leave existing services and cPanel-issued certificates intact. Keep new deployment/configuration files for diagnosis unless the owner requests removal. Remove any separately approved staging firewall rule by its exact identity. Do not restore or modify unrelated server state.
