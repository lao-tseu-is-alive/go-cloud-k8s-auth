[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=your-github-account_template-4-your-project-name&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=your-github-account_template-4-your-project-name)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=your-github-account_template-4-your-project-name&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=your-github-account_template-4-your-project-name)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=your-github-account_template-4-your-project-name&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=your-github-account_template-4-your-project-name)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=your-github-account_template-4-your-project-name&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=your-github-account_template-4-your-project-name)
[![test](https://github.com/your-github-account/template-4-your-project-name/actions/workflows/test.yml/badge.svg)](https://github.com/your-github-account/template-4-your-project-name/actions/workflows/test.yml)
[![cve-trivy-scan](https://github.com/your-github-account/template-4-your-project-name/actions/workflows/cve-trivy-scan.yml/badge.svg)](https://github.com/your-github-account/template-4-your-project-name/actions/workflows/cve-trivy-scan.yml)
[![codecov](https://codecov.io/gh/your-github-account/template-4-your-project-name/branch/main/graph/badge.svg?token=02AHW79CES)](https://codecov.io/gh/your-github-account/template-4-your-project-name)

# 🚀 template-4-your-project-name

A modern **Proto-first** microservice for managing "template4YourProjectNames" — built with Go, gRPC, ConnectRPC, and designed for cloud-native Kubernetes deployments.

> **Proto as Source of Truth**: API contracts are defined in Protocol Buffers, generating both Go code and OpenAPI specs automatically. Clients can connect via REST, gRPC, or Connect protocols.

## ✨ Features

- 🔐 **JWT Authentication** — Secure endpoints with token-based auth from [go-cloud-k8s-user-group](https://github.com/lao-tseu-is-alive/go-cloud-k8s-user-group)
- 📡 **Multi-Protocol Support** — REST, gRPC, and Connect (JSON/Proto) via [Vanguard transcoding](https://github.com/connectrpc/vanguard-go)
- 📋 **Proto-First Design** — Single source of truth for API definitions
- 🐘 **PostgreSQL Backend** — Robust data persistence with pgx driver
- 🐳 **Container Ready** — Optimized Docker images with CVE scanning via Trivy
- ☸️ **Kubernetes Native** — Ready for K8s deployment with health checks and metrics

---

## 🏗️ Architecture

```mermaid
graph TB
    subgraph Clients["📱 Clients"]
        REST["🌐 REST<br/>GET /goapi/v1/template_4_your_project_name"]
        CONNECT["⚡ Connect<br/>JSON / Proto"]
        GRPC["🔌 gRPC"]
    end
    
    subgraph Server["🖥️ Echo Server"]
        VG["🔄 Vanguard Transcoder"]
        subgraph Services["Connect Services"]
            TS["template4YourProjectNameService"]
            TTS["Typetemplate4YourProjectNameService"]
        end
    end
    
    subgraph Core["⚙️ Business Layer"]
        BS["BusinessService"]
        ST["Storage Interface"]
    end
    
    subgraph Data["💾 Data Layer"]
        PG[(PostgreSQL)]
    end
    
    REST --> VG
    CONNECT --> VG
    GRPC --> VG
    VG --> TS
    VG --> TTS
    TS --> BS
    TTS --> BS
    BS --> ST
    ST --> PG
```

---

## 📦 Proto-First API Design

The API is defined using **Protocol Buffers** as the single source of truth:

```
api/proto/template_4_your_project_name/v1/
├── template_4_your_project_name.proto           # template4YourProjectNameService definitions
└── type_template_4_your_project_name.proto      # Typetemplate4YourProjectNameService definitions
```

### Generated Artifacts

| Source | Generated | Purpose |
|--------|-----------|---------|
| `.proto` files | `gen/template_4_your_project_name/v1/*.go` | Go types & gRPC stubs |
| `.proto` files | `gen/template_4_your_project_name/v1/template_4_your_project_namev1connect/*.go` | Connect handlers |
| `.proto` files | `api/openapi/template_4_your_project_name.yaml` | OpenAPI 3.0 spec |

### Regenerate Code

```bash
./scripts/buf_generate.sh
# or
buf generate api/proto
```

---

## 🔌 API Endpoints

All endpoints are prefixed with `/goapi/v1` and require JWT authentication.

### template4YourProjectName Resources

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/goapi/v1/template_4_your_project_name` | List template_4_your_project_names |
| `POST` | `/goapi/v1/template_4_your_project_name` | Create a template_4_your_project_name |
| `GET` | `/goapi/v1/template_4_your_project_name/{id}` | Get template_4_your_project_name by ID |
| `PUT` | `/goapi/v1/template_4_your_project_name/{id}` | Update a template_4_your_project_name |
| `DELETE` | `/goapi/v1/template_4_your_project_name/{id}` | Delete a template_4_your_project_name |
| `GET` | `/goapi/v1/template_4_your_project_name/search` | Search template_4_your_project_names |
| `GET` | `/goapi/v1/template_4_your_project_name/count` | Count template_4_your_project_names |
| `GET` | `/goapi/v1/template_4_your_project_name/geojson` | Get GeoJSON |

### Typetemplate4YourProjectName Resources

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/goapi/v1/types` | List type template_4_your_project_names |
| `POST` | `/goapi/v1/types` | Create type template_4_your_project_name |
| `GET` | `/goapi/v1/types/{id}` | Get type template_4_your_project_name by ID |
| `PUT` | `/goapi/v1/types/{id}` | Update type template_4_your_project_name |
| `DELETE` | `/goapi/v1/types/{id}` | Delete type template_4_your_project_name |
| `GET` | `/goapi/v1/types/count` | Count type template_4_your_project_names |

### Connect RPC Endpoints

```bash
# Connect JSON format
curl -X POST http://localhost:9090/template_4_your_project_name.v1.template4YourProjectNameService/List \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"limit": 10}'
```

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- [buf](https://buf.build/docs/installation) (for proto generation)

### Environment Variables

```bash
# Required
export PORT=9090
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=template_4_your_project_name
export DB_USER=your_user
export DB_PASSWORD=your_password
export JWT_SECRET=your_jwt_secret
export ADMIN_PASSWORD=your_admin_password
```

### Run Locally

```bash
# Install dependencies
go mod download

# Run database migrations
# (migrations are auto-applied on startup)

# Start the server
go run ./cmd/template4YourProjectNameServer
```

### Run Tests

```bash
make test
```

---

## 🐳 Docker

### Pull from GitHub Container Registry

```bash
docker pull ghcr.io/your-github-account/template-4-your-project-name:latest
```

### Build Locally

```bash
docker build -t template-4-your-project-name .
```

Find all available versions in the [Packages section](https://github.com/your-github-account/template-4-your-project-name/pkgs/container/template-4-your-project-name).

---

## 📚 Documentation

- 📋 [Requirements](./documentation/Requirements.md) — Functional and system requirements
- 🔗 [OpenAPI Spec (YAML)](./api/openapi/template_4_your_project_name.yaml) — Generated from proto

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
template-4-your-project-name/
├── api/
│   ├── proto/template_4_your_project_name/v1/          # 📋 Proto definitions (source of truth)
│   └── openapi/                  # 📄 Generated OpenAPI specs
├── cmd/
│   └── template4YourProjectNameServer/   # 🚀 Main application entry point
├── gen/
│   └── template_4_your_project_name/v1/                # ⚙️ Generated Go code from protos
├── pkg/
│   └── template_4_your_project_name/                   # 📦 Business logic
│       ├── business_service.go  # Core business operations
│       ├── connect_server.go    # Connect RPC handlers
│       ├── mappers.go           # Domain ↔ Proto conversion
│       └── storage_postgres.go  # Database operations
├── db/migrations/               # 🗃️ SQL migrations
├── scripts/                     # 🔧 Build & generation scripts
└── documentation/               # 📚 Requirements & docs
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
