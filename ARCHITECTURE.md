# Poster Generator - Architecture

## Request Flow

```
┌──────────┐     ┌─────────────┐     ┌────────────────┐     ┌──────────────┐
│  Client  │────▶│ API Gateway │────▶│    Lambda      │────▶│ Avatar URL   │
│          │◀────│  (HTTP API) │◀────│  (Go/ARM64)    │◀────│ (External)   │
└──────────┘     └─────────────┘     └────────────────┘     └──────────────┘
                                            │
                                            ▼
                                     ┌──────────────┐
                                     │ Embedded     │
                                     │ Assets       │
                                     │ - font.ttf   │
                                     │ - image.jpg  │
                                     └──────────────┘
```

## Components

| Component | Purpose |
|-----------|---------|
| API Gateway | HTTP endpoint, CORS, routing |
| Lambda | Image processing, Go runtime |
| Embedded Assets | Font + background bundled in binary |

## Request Lifecycle

```
1. POST /generate
   {"name": "John", "avatarUrl": "https://..."}
        │
        ▼
2. API Gateway validates & forwards to Lambda
        │
        ▼
3. Lambda (cold start if new instance)
   - Parse JSON request
   - Fetch avatar from URL
        │
        ▼
4. Image Processing
   - Load embedded background (picture.jpg)
   - Load embedded font (poppins.ttf)
   - Resize avatar to 150x150
   - Draw avatar at center
   - Render name text below avatar
   - Encode as PNG
        │
        ▼
5. Response
   {"image": "base64..."}
```

## Project Structure

```
poster/
├── cmd/lambda/main.go        # Entry point, request handling
├── internal/poster/
│   ├── generator.go          # Image processing logic
│   └── assets/               # Embedded at compile time
│       ├── picture.jpg
│       └── poppins.ttf
└── infra/main.go             # CDK infrastructure
```

## Key Design Decisions

1. **Embedded Assets** - Font and background compiled into binary via `//go:embed`. No S3 or file system needed at runtime.

2. **ARM64** - Graviton2 processors for better price/performance.

3. **HTTP API** - Simpler and cheaper than REST API Gateway.

4. **Base64 Response** - Direct image return without intermediate storage.

5. **Stateless** - Each request is independent. Lambda scales horizontally.

## Cold vs Warm Start

```
Cold Start (new instance):
  Load binary → Init runtime → Load embedded assets → Process
  ~2-5s additional latency

Warm Start (reused instance):
  Process request immediately
  ~400-900ms latency
```

## Scaling Behavior

```
Low traffic:     1 instance
                 ┌───┐
                 │ λ │
                 └───┘

High traffic:    Auto-scales to N instances
                 ┌───┐ ┌───┐ ┌───┐ ┌───┐
                 │ λ │ │ λ │ │ λ │ │ λ │
                 └───┘ └───┘ └───┘ └───┘
```

## Deployment

```
make build  →  Compiles Go binary (linux/arm64)
make deploy →  CDK creates/updates CloudFormation stack
```

CDK provisions:
- Lambda function
- API Gateway HTTP API
- IAM execution role
- CloudWatch log group
