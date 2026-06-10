[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-auth&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-auth)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-auth&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-auth)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-auth&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-auth)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-auth&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-auth)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-auth&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-auth)
[![test](https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/actions/workflows/test.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/actions/workflows/test.yml)
[![cve-trivy-scan](https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/actions/workflows/cve-trivy-scan.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/actions/workflows/cve-trivy-scan.yml)
[![codecov](https://codecov.io/gh/lao-tseu-is-alive_/go-cloud-k8s-auth/branch/main/graph/badge.svg?token=02AHW79CES)](https://codecov.io/gh/lao-tseu-is-alive_/go-cloud-k8s-auth)


https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common-libs

# 🚀 go-cloud-auth

A modern **Proto-first** microservice for managing users and authentication — built with Go, gRPC, ConnectRPC, and designed for cloud-native Kubernetes deployments.

> **Proto as Source of Truth**: API contracts are defined in Protocol Buffers, generating both Go code and OpenAPI specs automatically. Clients can connect via REST, gRPC, or Connect protocols.

## ✨ Features

- 🔐 **OAuth2 Social Login** — Google, GitHub, and Microsoft AD authentication
- 🍪 **Browser SSO for module apps** — hosted login page, DB-backed session cookie, and a silent JWT mint endpoint so any module app (e.g. [go-mcp-markdown-notes](https://github.com/lao-tseu-is-alive/go-mcp-markdown-notes)) gets login with zero token copy-paste
- 🎟️ **Personal Access Tokens (PAT)** — long-lived revocable `pat_...` tokens for programmatic clients (MCP servers, scripts), with a self-service management UI at `/tokens.html` and a public introspection endpoint
- 👤 **User Management** — Auto-upsert profiles on login and standard CRUD support
- 🔑 **JWT Signing & Verification** — Cryptographically signs and validates user session tokens
- 🛡️ **Role-Based Access Control (RBAC)** — Simple role checks (admin, user) with group support
- 📡 **Multi-Protocol Support** — REST, gRPC, and Connect (JSON/Proto) via [Vanguard transcoding](https://github.com/connectrpc/vanguard-go)
- 🐘 **PostgreSQL Backend** — Robust data persistence with pgx driver and database migrations
- 🐳 **Container Ready** — Optimized Docker images with CVE scanning via Trivy
- ☸️ **Kubernetes Native** — Ready for K8s deployment with health checks, metrics, and Prometheus support

---

## 🏗️ Architecture

```mermaid
graph TB
    subgraph Clients["📱 Clients"]
        REST["🌐 REST<br/>POST /goapi/v1/auth/..."]
        CONNECT["⚡ Connect<br/>JSON / Proto"]
        GRPC["🔌 gRPC"]
    end
    
    subgraph Server["🖥️ Echo Server"]
        VG["🔄 Vanguard Transcoder"]
        subgraph Services["Connect Services"]
            AS["AuthService"]
            US["UserService"]
        end
    end
    
    subgraph Core["⚙️ Business Layer"]
        ABS["AuthBusinessService"]
        UBS["UserBusinessService"]
        ST["UserStorage (PostgreSQL)"]
    end
    
    subgraph Data["💾 Data Layer"]
        PG[(PostgreSQL)]
    end
    
    REST --> VG
    CONNECT --> VG
    GRPC --> VG
    VG --> AS
    VG --> US
    AS --> ABS
    US --> UBS
    ABS --> ST
    UBS --> ST
    ST --> PG
```

---

## 📦 Proto-First API Design

The API is defined using **Protocol Buffers** as the single source of truth:

```
proto/auth/v1/
└── auth.proto                 # AuthService & UserService definitions
```

### Generated Artifacts

| Source | Generated | Purpose |
|--------|-----------|---------|
| `.proto` files | `gen/auth/v1/*.go` | Go types & gRPC stubs |
| `.proto` files | `gen/auth/v1/authv1connect/*.go` | Connect handlers |
| `.proto` files | `api/openapi/auth.swagger.yaml` | OpenAPI 2.0 spec |

### Regenerate Code

```bash
./scripts/buf_generate.sh
# or
buf generate
```

---

## 🔌 API Endpoints

All endpoints are prefixed with `/goapi/v1` and require JWT authentication (except public OAuth endpoints: `StartOAuth`, `OAuthCallback`, and `ValidateToken`).

### Authentication Resources (AuthService)

| Method | Endpoint | Description | Public / Secured |
|--------|----------|-------------|------------------|
| `POST` | `/goapi/v1/auth/start` | Initiates OAuth flow (Google, GitHub, Microsoft) | **Public** |
| `POST` | `/goapi/v1/auth/callback` | Exhanges authorization code for signed JWT | **Public** |
| `POST` | `/goapi/v1/auth/validateToken` | Verifies a JWT token's validity | **Public** |
| `POST` | `/goapi/v1/auth/introspect` | Verifies a `pat_...` personal access token, returns identity + scopes | **Public** (the PAT is the credential) |
| `GET` | `/goapi/v1/auth/currentUser` | Returns details of the currently logged-in user | **Secured** |
| `POST` | `/goapi/v1/auth/tokens` | Create a personal access token (value returned **once**) | **Secured** |
| `GET` | `/goapi/v1/auth/tokens` | List the current user's PATs (metadata only) | **Secured** |
| `DELETE` | `/goapi/v1/auth/tokens/{id}` | Revoke one of the current user's PATs | **Secured** |

### Browser SSO Endpoints (plain Echo, cookie-based)

These endpoints implement the cross-module SSO flow. Anything involving the
session cookie or a browser redirect lives here (not in Connect RPC).

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/auth/login?redirect_uri=...` | Hosted login page (provider buttons). Redirects straight back if a valid session cookie exists. The `redirect_uri` must match the `ALLOWED_REDIRECT_URIS` allowlist. |
| `GET` | `/auth/oauth/{provider}/start?redirect_uri=...` | Starts the OAuth dance (302 to the provider consent page) |
| `GET` | `/auth/oauth/{provider}/callback` | Provider callback: creates the DB-backed session, sets the `goSession` HttpOnly cookie, 302 back to the module |
| `GET` | `/auth/token` | **Silent token mint**: exchanges the session cookie for a fresh short-lived JWT (`{token, expires_in_seconds, user}`). Returns 401 when no valid session. Callable cross-origin with `credentials: 'include'` from `ALLOWED_ORIGINS`. |
| `POST` | `/auth/logout` | Revokes the session and clears the cookie |
| `GET` | `/tokens.html` | Self-service PAT management UI (create / list / revoke) |

### User Resources (UserService)

| Method | Endpoint | Description | Public / Secured |
|--------|----------|-------------|------------------|
| `GET` | `/goapi/v1/user` | List users with pagination | **Secured** |
| `POST` | `/goapi/v1/user` | Create a user | **Secured (Admin)** |
| `GET` | `/goapi/v1/user/{id}` | Get user details by UUID | **Secured** |
| `PUT` | `/goapi/v1/user/{id}` | Update a user's details | **Secured (Admin)** |
| `DELETE` | `/goapi/v1/user/{id}` | Delete a user by UUID | **Secured (Admin)** |
| `GET` | `/goapi/v1/user/count` | Count total users in the DB | **Secured** |
| `GET` | `/goapi/v1/user/by-external-id/{external_id}` | Retrieve a user by their legacy integer ID | **Secured** |

### Connect RPC Endpoints

```bash
# Connect JSON format
curl -X POST http://localhost:9090/auth.v1.UserService/List \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"limit": 10}'
```

---

## 🍪 Browser SSO: how module apps log users in

The copy-paste-a-JWT era is over. A module app (running on the same parent
domain in production, or another `localhost` port in dev — cookies are **not**
port-scoped) integrates like this:

```
Browser                    module app (:8080)              auth service (:9090)
   │  no token? ───────────────► GET /auth/token (cookie) ──► 401
   │  window.location = :9090/auth/login?redirect_uri=http://localhost:8080/
   │  ── pick Google/GitHub ──► OAuth dance ──► Set-Cookie: goSession ──► 302 back
   │  GET /auth/token (cookie, credentials:'include') ──► { token, expires_in_seconds, user }
   │  use JWT in memory, re-mint silently at ~80% of its lifetime
```

A new module needs only three things:
1. Verify JWTs with the shared `JWT_SECRET` (same as today).
2. Be listed in `ALLOWED_ORIGINS` (CORS w/ credentials) and `ALLOWED_REDIRECT_URIS`.
3. ~20 lines of frontend: redirect to `/auth/login` when `/auth/token` returns 401,
   otherwise keep the minted JWT in memory only.

Sessions are stored in `go_auth.sessions` (opaque cookie value, SHA-256 hash in
DB) so logout and admin revocation work instantly. Personal access tokens
(`go_auth.personal_access_tokens`) cover non-browser clients; downstream
services verify them via `POST /goapi/v1/auth/introspect`.

> 📋 Setting this up end-to-end (OAuth consoles, env vars, first login) is
> described step by step in [documentation/sso_setup_checklist.md](./documentation/sso_setup_checklist.md).

---

## 🚀 Configuration & Testing Guide

This section explains how to configure, run, and test the `go-cloud-k8s-auth` authentication and user management service.

### 📋 Prerequisites
- **Go 1.25+**
- **PostgreSQL 14+** (with a database named `go_cloud_auth`)
- **buf** CLI (to regenerate proto files if you edit them) [The Buf CLI home page](https://buf.build/product/cli)

---

### ⚙️ Step 1: Environment Configuration

Create a `.env` file at the root of the project by copying [.env_sample](file:///home/cgil/cgdev/golang/go-cloud-k8s-auth/.env_sample):
```bash
cp .env_sample .env
```

Adapt the variables to your setup. For OAuth2 providers, here is how to register an application to get credentials:

#### GitHub OAuth Setup
1. Go to **Settings > Developer Settings > OAuth Apps > New OAuth App** on GitHub.
2. Set **Homepage URL** to `http://localhost:9090`.
3. Set the **Authorization callback URL(s)**. Two flows exist, each with its own callback:
   - **Browser SSO flow** (recommended, used by module apps): `http://localhost:9090/auth/oauth/github/callback`
   - Legacy SPA flow (this service's own demo page): `http://localhost:9090/`

   GitHub OAuth Apps accept a single callback URL: prefer the SSO one (or create
   one OAuth App per flow). Google and Microsoft accept multiple redirect URIs,
   so register both there (`/auth/oauth/google/callback`, `/auth/oauth/microsoft/callback`).
4. Register the app, generate a Client Secret, and configure `.env`:
   ```env
   # Google
   OAUTH_GOOGLE_CLIENT_ID="your_google_client_id"
   OAUTH_GOOGLE_CLIENT_SECRET="your_google_client_secret"
   OAUTH_GOOGLE_REDIRECT_URL="http://localhost:9090/"

   # GitHub
   OAUTH_GITHUB_CLIENT_ID="your_github_client_id"
   OAUTH_GITHUB_CLIENT_SECRET="your_github_client_secret"
   OAUTH_GITHUB_REDIRECT_URL="http://localhost:9090/"

   # Microsoft Azure AD
   OAUTH_MICROSOFT_CLIENT_ID="your_microsoft_client_id"
   OAUTH_MICROSOFT_CLIENT_SECRET="your_microsoft_client_secret"
   OAUTH_MICROSOFT_REDIRECT_URL="http://localhost:9090/"
   ```
   *Note:* `OAUTH_*_REDIRECT_URL` only configures the legacy SPA flow; the
   browser SSO flow always derives its callback from `AUTH_PUBLIC_BASE_URL`
   (`<AUTH_PUBLIC_BASE_URL>/auth/oauth/<provider>/callback`).

#### Browser SSO variables

| Variable | Dev default | Purpose |
|----------|-------------|---------|
| `AUTH_PUBLIC_BASE_URL` | `http://localhost:9090` | Externally reachable base URL, used to build the OAuth callback URLs |
| `SESSION_COOKIE_NAME` | `goSession` | Name of the HttpOnly session cookie |
| `COOKIE_DOMAIN` | *(empty = host-only)* | Cookie `Domain` attribute; set the parent domain (e.g. `.example.com`) in production |
| `COOKIE_SECURE` | `false` | Set `true` behind https |
| `SESSION_DURATION` | `720h` | Browser session lifetime (Go duration) |
| `ALLOWED_REDIRECT_URIS` | `http://localhost:8080` | Comma-separated URL prefixes allowed as `redirect_uri` |
| `ALLOWED_ORIGINS` | `https://golux.lausanne.ch,http://localhost:3000,http://localhost:8080` | CORS origins allowed to call `/auth/token` with credentials |

---

### 🚀 Step 2: Start the Server

1. Install dependencies:
   ```bash
   go mod download
   ```
2. Start the server (migrations will auto-apply to the configured PostgreSQL database on startup):
   ```bash
   go run ./cmd/goCloudAuthServer
   ```
   *Alternatively, you can run `make run` if your environment has the variables exported.*

---

### 🧪 Step 3: Testing the OAuth2 Authentication Flow

Since OAuth2 requires browser interaction to authenticate the user, you can test the flow using `curl` and your browser.

#### 1. Initiate the OAuth Flow
Ask the service to generate the authorization URL for your chosen provider (e.g., `github`):
```bash
curl -X POST http://localhost:9090/goapi/v1/auth/start \
  -H "Content-Type: application/json" \
  -d '{"provider": "github"}'
```
**Response**:
```json
{
  "authUrl": "https://github.com/login/oauth/authorize?client_id=...&redirect_uri=...&state=...",
  "state": "random_anti_csrf_token_here"
}
```

#### 2. Authenticate in Browser
Copy the `authUrl` from the response and paste it into your browser. 
Once you log in and authorize the app on GitHub, your browser will redirect to the callback URL, which will look like:
`http://localhost:9090/goapi/v1/auth/callback?code=AUTH_CODE&state=STATE_TOKEN`

#### 3. Complete the Callback
Vanguard transcodes this callback automatically. Exchange the `code` and `state` to retrieve your signed JWT and user info:
```bash
curl -X POST http://localhost:9090/goapi/v1/auth/callback \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "github",
    "code": "AUTH_CODE_FROM_URL",
    "state": "STATE_TOKEN_FROM_URL"
  }'
```
**Response**:
```json
{
  "jwt": "eyJhbGciOi...",
  "user": {
    "id": "user-uuid-v4-here",
    "externalId": 1234567,
    "email": "user@example.com",
    "name": "User Name",
    "avatarUrl": "https://avatars.githubusercontent.com/...",
    "provider": "github",
    "roles": ["user"]
  }
}
```
*Note: The user is automatically upserted (created or updated) in PostgreSQL during this callback.*

---

### 🔐 Step 4: Accessing Secured API Endpoints

Once you have your JWT token, store it in a variable:
```bash
export JWT_TOKEN="eyJhbGciOi..."
```

#### 1. Get Current User info
```bash
curl -H "Authorization: Bearer $JWT_TOKEN" \
  http://localhost:9090/goapi/v1/auth/currentUser
```

#### 2. Validate the Token
```bash
curl -X POST http://localhost:9090/goapi/v1/auth/validateToken \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$JWT_TOKEN\"}"
```

#### 3. List Users (Requires JWT)
```bash
curl -H "Authorization: Bearer $JWT_TOKEN" \
  http://localhost:9090/goapi/v1/user
```

#### 4. Create User (Admin required or user creation)
```bash
curl -X POST http://localhost:9090/goapi/v1/user \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user": {
      "email": "test-user@example.org",
      "name": "Test User",
      "roles": ["user"]
    }
  }'
```

---

## 💻 Example ConnectRPC Client

We provide a fully functional example ConnectRPC client in [cmd/exampleClient](file:///home/cgil/cgdev/golang/go-cloud-k8s-auth/cmd/exampleClient/main.go) to demonstrate how your cloud-native services can connect to and query this auth microservice in Go.

The client demonstrates:
1. ConnectRPC client instantiation for both `AuthService` and `UserService`.
2. Setting up a client-side unary interceptor to inject a JWT Bearer token into outgoing requests.
3. Performing token validation, retrieving current user profile, and listing users.

### Build the Example Client

```bash
make build-example-client
```

### Run the Example Client

Ensure the server is running (e.g. `go run ./cmd/goCloudAuthServer`), then run the client commands:

#### 1. Initiate OAuth Login Flow
Generates a login link for a provider (defaults to `github`):
```bash
./bin/exampleClient -cmd start-oauth -provider github
# or via make
make run-example-client ARGS="-cmd start-oauth -provider github"
```

#### 2. Validate a JWT Token
```bash
./bin/exampleClient -cmd validate -token "YOUR_JWT_TOKEN"
```

#### 3. Fetch Currently Logged-in User Profile
This requires the token to be sent in the request authorization headers via client interceptor:
```bash
./bin/exampleClient -cmd current-user -token "YOUR_JWT_TOKEN"
```

#### 4. List Registered Users
Queries the user service (requires admin rights or appropriate JWT token roles):
```bash
./bin/exampleClient -cmd list-users -token "YOUR_JWT_TOKEN" -limit 5
```

---

### 🧪 Run Automated Tests

```bash
make test
```

---

## 🐳 Docker

### Pull from GitHub Container Registry

```bash
docker pull ghcr.io/lao-tseu-is-alive/go-cloud-k8s-auth:latest
```

### Build Locally

```bash
docker build -t go-cloud-k8s-auth .
```

Find all available versions in the [Packages section](https://github.com/lao-tseu-is-alive/go-cloud-k8s-auth/pkgs/container/go-cloud-k8s-auth).

---

## 📚 Documentation

- 📋 [Requirements](./documentation/Requirements.md) — Functional and system requirements
- 🔐 [OAuth2 Providers Setup](./documentation/oauth_providers_setup.md) — Detailed guide to configure Google, GitHub, and Microsoft credentials
- 🍪 [SSO Setup Checklist](./documentation/sso_setup_checklist.md) — **Step-by-step manual actions** to bring the browser SSO + PAT + MCP chain online
- 🔗 [OpenAPI Spec (YAML)](./api/openapi/go_cloud_auth.yaml) — Generated from proto

---

## 🛠️ Tech Stack

| Category | Technology |
|----------|------------|
| **Language** | Go 1.21+ |
| **API Framework** | [Echo](https://echo.labstack.com/) |
| **RPC** | [ConnectRPC](https://connectrpc.com/) + [Vanguard](https://github.com/connectrpc/vanguard-go) |
| **Proto Tooling** | [buf](https://buf.build/) |
| **Database** | PostgreSQL with [pgx](https://github.com/jackc/pgx) |
| **Auth** | JWT via [cristalhq/jwt](https://github.com/cristalhq/jwt) |
| **Monitoring** | Prometheus metrics |
| **Container** | Docker with multi-stage builds |
| **Security** | [Trivy](https://aquasecurity.github.io/trivy/) CVE scanning |

---

## 📁 Project Structure

```
go-cloud-k8s-auth/
├── proto/
│   └── auth/v1/                 # 📋 Proto definitions (source of truth)
├── api/
│   └── openapi/                 # 📄 Generated OpenAPI specs
├── cmd/
│   └── goCloudAuthServer/       # 🚀 Main application entry point
├── gen/
│   └── auth/v1/                 # ⚙️ Generated Go code from protos
├── pkg/
│   ├── auth/                    # 📦 Authentication & User services
│   │   ├── auth_service.go      # OAuth and token business logic
│   │   ├── user_service.go      # User CRUD business logic
│   │   ├── auth_connect_server.go # ConnectRPC auth handlers
│   │   ├── user_connect_server.go # ConnectRPC user handlers
│   │   ├── browser_handlers.go  # Browser SSO: /auth/login, OAuth callbacks, /auth/token, /auth/logout
│   │   ├── session_store.go     # DB-backed browser sessions (go_auth.sessions)
│   │   ├── pat_service.go       # Personal Access Tokens: generation, introspection
│   │   ├── pat_store.go         # PAT persistence (go_auth.personal_access_tokens)
│   │   ├── pat_connect_server.go # ConnectRPC PAT handlers (introspect + CRUD)
│   │   ├── storage.go           # Storage interfaces
│   │   ├── storage_postgres.go  # PostgreSQL operations (pgx)
│   │   ├── auth_interceptor.go  # JWT validation interceptor
│   │   ├── mappers.go           # Domain ↔ Proto conversion stubs
│   │   ├── state_store.go       # Safe in-memory anti-CSRF store
│   │   └── errors.go / messages.go # Domain validation & error constants
│   └── rbac/                    # 🛡️ Role-Based Access Control
│       ├── enforcer.go          # RBAC interface
│       └── simple_enforcer.go   # Admin & User static rules
├── db/migrations/               # 🗃️ SQL migrations
└── scripts/                     # 🔧 Build & generation scripts
```

---

## 🔒 Security

- All CVE scans performed automatically before container builds
- JWT authentication required for all `/goapi/v1/*` endpoints
- SonarCloud analysis for code quality and security
- Dependabot for dependency updates

---

## 📄 License

MIT License — See [LICENSE](./LICENSE) for details.

---

<p align="center">
  Built with ❤️ using Go, Proto, and Connect
</p>
