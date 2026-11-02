# Password reset

The login and admin-login pages request a reset link by email. Passwords are changed
only on `/reset-password` after the user opens that link.

- Random 256-bit tokens; only SHA-256 hashes are stored in `users.reset_token`.
- Database expiry: 15 minutes. A new link replaces the previous link.
- Password update and token deletion are one atomic SQL statement.
- Requests for the same account have a persistent 60-second cooldown.
- Requests are limited per client IP per server instance: 5 reset emails per
  15 minutes and 10 reset submissions per minute. Shared gateways must enforce
  an additional global limit when running multiple instances.
- Responses never reveal whether an account exists or return its reset token.
- A bounded background worker sends via Brevo. Delivery failures are logged and
  revoke only the failed token, permitting another request. API acceptance does
  not guarantee inbox delivery; inspect Brevo delivery events for that.
- Password changes also clear any outstanding reset token.
- Existing login JWTs retain their current lifetime; this change does not add
  session revocation.

## Deployment

Set these environment variables on the backend:

```dotenv
APP_URL=https://your-frontend-domain.example
BREVO_API_KEY=your-brevo-api-key
EMAIL_SENDER=your-verified-sender@example.com
```

HTTP is accepted only for localhost development. The sender must be authorized in
Brevo. The existing `users.reset_token` and `users.reset_token_expiry` columns are
required; they are already referenced by the user model and repository.

If deployed behind a reverse proxy, set `TRUSTED_PROXIES` to its actual trusted
IP addresses/CIDRs (comma separated). By default forwarded IP headers are ignored.
Do not trust all networks: otherwise clients can spoof the rate-limit key.

Previously issued JWT reset links stop working; users must request new links.

## Checks

```shell
go test ./internal/utils ./internal/middleware ./internal/repository ./internal/service ./internal/handler ./internal/routes
```

PostgreSQL lifecycle tests are opt-in: set `TEST_DATABASE_URL` to a disposable
test database, then run `go test ./internal/repository -run TestPasswordResetLifecycle`.
The default run does not contact a database or send real emails.

For a read-only live authentication check, run `go run ./cmd/check-brevo` from
the backend root. This loads environment variables the same way as the server,
checks `GET /v3/account`, and prints no API keys or account details. It sends no
email and does not contact PostgreSQL. A 401 may be an unrecognized outbound IP,
not just an invalid key; the diagnostic reports Brevo's error message. Review
the IP yourself in Brevo's Authorized IPs settings before authorizing it.
