.PHONY: build deploy destroy clean deps

# AWS Profile
PROFILE = techpix
AWS_ARGS = --profile $(PROFILE)

# Build the Lambda binary
build:
	@echo "Building Lambda binary..."
	@mkdir -p build
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o build/bootstrap cmd/lambda/main.go
	@echo "Build complete: build/bootstrap"

# Install dependencies
deps:
	@echo "Installing Lambda dependencies..."
	go mod tidy
	@echo "Installing CDK dependencies..."
	cd infra && go mod tidy

# Deploy the stack
deploy: build
	@echo "Deploying CDK stack..."
	cd infra && cdk deploy --require-approval never $(AWS_ARGS)

# Destroy the stack
destroy:
	@echo "Destroying CDK stack..."
	cd infra && cdk destroy --force $(AWS_ARGS)

# Bootstrap CDK (run once per account/region)
bootstrap:
	@echo "Bootstrapping CDK..."
	cd infra && cdk bootstrap $(AWS_ARGS)

# Synthesize CloudFormation template
synth:
	cd infra && cdk synth $(AWS_ARGS)

# Clean build artifacts
clean:
	rm -rf build
	rm -rf infra/cdk.out

# Local test (requires assets to be present)
test-local:
	@echo "Running local test..."
	go run cmd/lambda/main.go


image:
	cat response.json | jq -r '.image' | base64 -d > output.png && open output.png