# Poster Generator - Code Walkthrough

## File Overview

| File | Purpose | Lines |
|------|---------|-------|
| `cmd/lambda/main.go` | Lambda entry point, HTTP handling | 82 |
| `internal/poster/generator.go` | Image processing engine | 155 |
| `infra/main.go` | AWS infrastructure (CDK) | 88 |

---

## 1. Lambda Entry Point

**File:** `cmd/lambda/main.go`

### Initialization (runs once per cold start)

```go
// Line 25-33
var generator *poster.Generator

func init() {
    var err error
    generator, err = poster.NewGenerator()  // Loads font + background
    if err != nil {
        log.Fatalf("failed to initialize generator: %v", err)
    }
}
```

- `init()` runs **once** when Lambda cold starts
- Creates `Generator` instance with embedded assets
- Stored in package-level variable for reuse across warm invocations

### Request/Response Types

```go
// Line 14-23
type Request struct {
    Name      string `json:"name"`
    AvatarURL string `json:"avatarUrl"`
}

type Response struct {
    Image string `json:"image,omitempty"`
    Error string `json:"error,omitempty"`
}
```

### Handler Function

```go
// Line 35-65
func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
```

**Flow:**
1. `Line 36-38` - Parse JSON body into `Request` struct
2. `Line 41-47` - Validate required fields
3. `Line 49` - Call `generator.Generate()` to create image
4. `Line 55-64` - Return base64 image in JSON response

### Lambda Start

```go
// Line 80-82
func main() {
    lambda.Start(handler)
}
```

- `lambda.Start()` registers handler with AWS Lambda runtime
- Runtime calls `handler()` for each incoming request

---

## 2. Image Processing Engine

**File:** `internal/poster/generator.go`

### Asset Embedding

```go
// Line 21-23
//go:embed assets/picture.jpg
//go:embed assets/poppins.ttf
var Assets embed.FS
```

- `//go:embed` directive compiles files into binary
- No filesystem access needed at runtime
- Assets loaded from memory

### Constants

```go
// Line 25-29
const (
    avatarSize = 150  // Avatar dimensions in pixels
    fontSize   = 36   // Text size
    dpi        = 72   // Font rendering DPI
)
```

### Generator Struct

```go
// Line 32-35
type Generator struct {
    background image.Image    // Decoded background image
    font       *truetype.Font // Parsed TTF font
}
```

### NewGenerator - Asset Loading

```go
// Line 38-65
func NewGenerator() (*Generator, error)
```

**Steps:**
1. `Line 40` - Read background JPG from embedded filesystem
2. `Line 45` - Decode JPG to `image.Image`
3. `Line 51` - Read font TTF from embedded filesystem
4. `Line 56` - Parse TTF into usable font object

### Generate - Main Processing

```go
// Line 68-106
func (g *Generator) Generate(name, avatarURL string) (string, error)
```

**Step-by-step:**

| Line | Action |
|------|--------|
| 70 | Fetch avatar from URL via HTTP |
| 76 | Resize avatar to 150x150 using Lanczos3 algorithm |
| 79-80 | Create new RGBA image with background dimensions |
| 83 | Draw background onto output image |
| 87-88 | Calculate avatar X,Y for center positioning |
| 91-92 | Draw avatar onto output |
| 95 | Draw name text below avatar |
| 100-101 | Encode output as PNG |
| 105 | Convert to base64 string |

### Text Rendering

```go
// Line 109-130
func (g *Generator) drawText(img *image.RGBA, text string, y int) error
```

| Line | Action |
|------|--------|
| 110-117 | Configure freetype context (DPI, font, size, color) |
| 120-122 | Measure text width for centering |
| 125 | Calculate X position: `(imageWidth - textWidth) / 2` |
| 127-128 | Draw text at calculated position |

### Avatar Fetching

```go
// Line 133-155
func fetchImage(url string) (image.Image, error)
```

| Line | Action |
|------|--------|
| 134 | HTTP GET request to avatar URL |
| 140-141 | Validate HTTP 200 response |
| 144 | Read response body |
| 149 | Decode image (auto-detects PNG/JPEG) |

---

## 3. Infrastructure (CDK)

**File:** `infra/main.go`

### Lambda Definition

```go
// Line 24-34
fn := awslambda.NewFunction(stack, jsii.String("PosterGenerator"), &awslambda.FunctionProps{
    Runtime:      awslambda.Runtime_PROVIDED_AL2023(),  // Custom runtime
    Handler:      jsii.String("bootstrap"),              // Binary name
    Architecture: awslambda.Architecture_ARM_64(),       // Graviton2
    Code:         awslambda.Code_FromAsset(jsii.String("../build"), nil),
    MemorySize:   jsii.Number(512),                      // 512MB RAM
    Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
})
```

| Property | Value | Why |
|----------|-------|-----|
| Runtime | `PROVIDED_AL2023` | Custom Go binary, not managed runtime |
| Handler | `bootstrap` | AWS looks for this binary name |
| Architecture | `ARM_64` | Cheaper, faster than x86 |
| Code | `../build` | Directory containing compiled binary |
| MemorySize | `512` | Enough for image processing |
| Timeout | `30s` | Max time per request |

### API Gateway

```go
// Line 37-47
httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("PosterApi"), &awsapigatewayv2.HttpApiProps{
    ApiName: jsii.String("poster-api"),
    CorsPreflight: &awsapigatewayv2.CorsPreflightOptions{
        AllowOrigins: jsii.Strings("*"),
        AllowMethods: &[]awsapigatewayv2.CorsHttpMethod{
            awsapigatewayv2.CorsHttpMethod_POST,
            awsapigatewayv2.CorsHttpMethod_OPTIONS,
        },
        AllowHeaders: jsii.Strings("Content-Type"),
    },
})
```

- HTTP API (not REST API) - simpler, cheaper
- CORS enabled for browser requests

### Route Configuration

```go
// Line 57-61
httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
    Path:        jsii.String("/generate"),
    Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_POST},
    Integration: integration,
})
```

- `POST /generate` → Lambda function

### Output

```go
// Line 64-67
awscdk.NewCfnOutput(stack, jsii.String("ApiUrl"), &awscdk.CfnOutputProps{
    Value:       httpApi.Url(),
    Description: jsii.String("API Gateway URL"),
})
```

- Prints API URL after deployment

---

## Execution Flow (Line by Line)

```
Request arrives at API Gateway
        │
        ▼
cmd/lambda/main.go:35  →  handler() called
cmd/lambda/main.go:37  →  JSON parsed into Request
cmd/lambda/main.go:49  →  generator.Generate() called
        │
        ▼
internal/poster/generator.go:70   →  fetchImage(avatarURL)
internal/poster/generator.go:134  →  HTTP GET avatar
internal/poster/generator.go:149  →  Decode avatar image
        │
        ▼
internal/poster/generator.go:76   →  Resize avatar to 150x150
internal/poster/generator.go:83   →  Draw background
internal/poster/generator.go:92   →  Draw avatar on top
internal/poster/generator.go:95   →  drawText() called
internal/poster/generator.go:128  →  Render name text
        │
        ▼
internal/poster/generator.go:101  →  Encode as PNG
internal/poster/generator.go:105  →  Convert to base64
        │
        ▼
cmd/lambda/main.go:55-64  →  Return JSON response
```

---

## Key Libraries

| Import | Purpose | Used In |
|--------|---------|---------|
| `github.com/aws/aws-lambda-go` | Lambda runtime + events | `cmd/lambda/main.go` |
| `github.com/golang/freetype` | TrueType font rendering | `generator.go:110` |
| `github.com/nfnt/resize` | Image resizing | `generator.go:76` |
| `image/draw` | Image compositing | `generator.go:83,92` |
| `embed` | Compile-time asset bundling | `generator.go:23` |
