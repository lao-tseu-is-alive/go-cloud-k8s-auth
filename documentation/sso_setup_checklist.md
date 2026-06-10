# SSO + PAT + MCP — Manual Setup Checklist

The code for browser SSO, Personal Access Tokens and the `notes-mcp` server is
implemented and unit/e2e tested (June 2026). A few steps **cannot be automated**
because they involve third-party consoles and secrets. This page lists exactly
what a human must do, in order. Everything else (DB migrations, routes, UI) is
handled automatically at server startup.

Status legend: ☐ = action required, ✅ = nothing to do, already handled.

---

## 1. ☐ Register the new OAuth callback URLs (required once per provider)

The browser SSO flow uses dedicated callback URLs derived from
`AUTH_PUBLIC_BASE_URL`:

```
http://localhost:9090/auth/oauth/google/callback
http://localhost:9090/auth/oauth/github/callback
http://localhost:9090/auth/oauth/microsoft/callback
```

Until these are registered, the providers will refuse the SSO login with a
`redirect_uri mismatch` error.

- **Google** ([console.cloud.google.com](https://console.cloud.google.com/apis/credentials)):
  open your OAuth 2.0 Client ID → *Authorized redirect URIs* → **add**
  `http://localhost:9090/auth/oauth/google/callback` (keep the existing ones).
- **GitHub** ([github.com/settings/developers](https://github.com/settings/developers)):
  GitHub OAuth Apps accept **one** callback URL. Either change it to
  `http://localhost:9090/auth/oauth/github/callback` (the old SPA demo flow on
  `/` stops working), or create a second OAuth App dedicated to SSO.
- **Microsoft Entra ID** (Azure portal → App registrations → Authentication):
  add `http://localhost:9090/auth/oauth/microsoft/callback` as a Web redirect URI.

For production, register the same paths under your public base URL
(`https://auth.example.com/auth/oauth/<provider>/callback`) and set
`AUTH_PUBLIC_BASE_URL` accordingly.

## 2. ☐ Check the new env vars in `go-cloud-k8s-auth/.env`

Compare with `.env_sample` (section *BROWSER SSO CONFIGURATION*). The dev
defaults work out of the box for localhost; you only need to **add the
variables if you want non-default values**:

| Variable | Needs action? |
|----------|---------------|
| `AUTH_PUBLIC_BASE_URL` | only if the auth service is not on `http://localhost:9090` |
| `ALLOWED_REDIRECT_URIS` | only if a module app is not on `http://localhost:8080` |
| `ALLOWED_ORIGINS` | same — must contain each module app origin (CORS w/ credentials) |
| `COOKIE_DOMAIN` / `COOKIE_SECURE` | **production only**: parent domain (e.g. `.example.com`) + `true` |
| `SESSION_DURATION` / `SESSION_COOKIE_NAME` | optional tuning (defaults: `720h`, `goSession`) |

## 3. ✅ JWT secret shared with go-mcp-markdown-notes

`JWT_SECRET` must be identical in both `.env` files — **verified identical on
this machine** during e2e testing. Only act if you rotate the secret: rotate it
in both files.

## 4. ☐ Set `AUTH_SERVER_URL` in `go-mcp-markdown-notes/.env`

Add (see its `.env_sample`):

```env
NOTES_AUTH_MODE=jwt
AUTH_SERVER_URL=http://localhost:9090
```

## 5. ✅ Database migrations

Migrations `000002_create_sessions` and `000003_create_personal_access_tokens`
are embedded and **auto-applied at auth-server startup** (already applied to
the local dev DB). Nothing to run manually.

## 6. ☐ First end-to-end browser test (after step 1)

1. Start both servers (`go run ./cmd/goCloudAuthServer` and `make run` in the
   notes repo).
2. Open `http://localhost:8080/` in a fresh browser profile.
3. Click **Sign in with Google / GitHub** → provider login → you land back on
   the notes app, connected, and the notes list loads.
4. Sanity checks: DevTools → Application → no JWT in localStorage; the
   `goSession` cookie lives on `localhost` (set by :9090, sent to both ports —
   cookies are not port-scoped, this is expected).

## 7. ☐ Create a PAT and register notes-mcp in your MCP client

1. While signed in, click **Manage MCP tokens** in the notes app (or open
   `http://localhost:9090/tokens.html`).
2. Create a token (e.g. name `claude-mcp`, expiry 90 days). **Copy the
   `pat_...` value now — it is shown exactly once.**
3. Build and register the MCP server:
   ```bash
   cd go-mcp-markdown-notes && make build-mcp
   claude mcp add notes \
     -e NOTES_SERVER=http://127.0.0.1:8080 \
     -e NOTES_TOKEN=pat_...your_token... \
     -- "$PWD/bin/notes-mcp"
   ```
4. In Claude Code: *“liste mes notes récentes”*, *“crée une note titrée X”*,
   *“supprime cette note”*. Revoking the PAT in `tokens.html` cuts access
   within ≤ 60 seconds.

## 8. Known limitations / deferred follow-ups (decide later, nothing blocking)

- The OAuth **state store is in-memory**: in Kubernetes, run a single auth
  replica or enable sticky sessions, otherwise the OAuth dance may fail.
- The public `POST /goapi/v1/auth/introspect` endpoint has no rate limiting
  (PAT entropy ≈ 190 bits makes brute force impractical, but add limiting
  before internet exposure).
- Remote MCP over streamable HTTP + OAuth 2.1 was deliberately deferred
  (phase 2); the PAT design does not block it.
- The silent token mint requires auth service and modules to share a parent
  domain (or localhost). Different registrable domains would break it.
