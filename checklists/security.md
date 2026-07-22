# Security checklist

Walk this before merging security-sensitive changes and before any first external
deploy. Answer with evidence (a path, a test, a config), not intent. Items that do not
apply get "n/a — reason", not silence.

## Input and output

- [ ] All external input is validated at the trust boundary (type, length, range, format)
- [ ] Validation rejects by default; allowlists over denylists where practical
- [ ] Output is encoded for its context (HTML, SQL, shell, URL) — no string-built queries or commands
- [ ] File uploads are constrained (type, size) and stored outside the web root / executed never
- [ ] Error messages returned to clients carry no internals (stack traces, query text, paths)

## Authentication and sessions

- [ ] Passwords hashed with a current algorithm (argon2id, bcrypt at sane cost) — never stored or logged
- [ ] Tokens and session IDs are random, expire, and rotate on privilege change
- [ ] Session cookies: `HttpOnly`, `Secure`, `SameSite` set deliberately
- [ ] Failed auth attempts are rate-limited and logged without logging the credentials

## Authorisation

- [ ] Every handler checks authorisation server-side; no "hidden in the UI" checks
- [ ] Object-level access is checked (IDOR: user A cannot fetch user B's record by changing an ID)
- [ ] Default is deny; privileges are least-needed per role and per service account

## Secrets and configuration

- [ ] No secrets in the repo, in history, in logs, or in client-side code (grep the diff)
- [ ] Secrets come from a store or environment, not config files committed alongside code
- [ ] Dev and prod secrets differ; rotation is documented
- [ ] Debug endpoints and verbose modes are off in production config

## Data protection

- [ ] Sensitive data is identified; only what is needed is collected and kept
- [ ] TLS for data in transit; current protocol versions only
- [ ] Sensitive data encrypted at rest where the threat model calls for it
- [ ] Personal data (PHI, PII) never appears in logs, analytics, error reports, or prompts to external services
- [ ] Backups and exports get the same protection as the primary store

## Dependencies and supply chain

- [ ] Dependencies are pinned with a lockfile
- [ ] Known-vulnerability scanning runs in CI (Dependabot, `npm audit`, `govulncheck`, or equivalent)
- [ ] New dependencies are justified: maintained, licence-compatible, actually needed

## Web and API surface

- [ ] CORS allows known origins, not `*` with credentials
- [ ] Security headers set where relevant (CSP, `X-Content-Type-Options`, frame policy)
- [ ] Rate limiting on public endpoints, login, and anything expensive
- [ ] CSRF protection on state-changing requests that use cookies
- [ ] Mass assignment is blocked (explicit field allowlists on bind/update)

## Logging and detection

- [ ] Security-relevant events are logged: auth failures, privilege changes, denied access, admin actions
- [ ] Logs do not contain secrets, tokens, or personal data
- [ ] Alerts exist for the events that matter (repeated auth failure, unusual export volume)

## Before declaring done

- [ ] The threats considered are written down — one paragraph naming who attacks and what they want
- [ ] A second person (or a fresh session) re-checked authorisation and secrets items
