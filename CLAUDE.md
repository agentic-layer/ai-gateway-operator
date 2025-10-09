# AI Gateway Operator - Developer Guide

@README.md
@docs/modules/operator/partials/how-to-guide.adoc

## Overview

The AI Gateway Operator is a Kubernetes operator built with the Operator SDK framework that provides the base API infrastructure for managing AI gateway instances. This operator defines Custom Resource Definitions (CRDs) and webhooks for AI gateway management but **does not include controller implementations**. Actual gateway implementations are provided by separate operators (e.g., [ai-gateway-litellm-operator](https://github.com/agentic-layer/ai-gateway-litellm-operator)) that implement the controller logic for specific gateway technologies.

**Technology Stack:**
- **Framework**: Operator SDK v1.41+ (based on controller-runtime)
- **Language**: Go 1.24+
- **Kubernetes**: 1.33+ (uses admission webhooks, CRDs)
- **Pattern**: Base API operator (provides CRDs/webhooks, no controllers)

## Architecture Overview

### Core Components

1. **AiGateway CRD** (`api/v1alpha1/aigateway_types.go`)
   - Defines AI gateway instances with model routing configuration
   - Separate `name` and `provider` fields for AI models
   - Configurable port (default: 4000)
   - References AiGatewayClass to select implementation

2. **AiGatewayClass CRD** (`api/v1alpha1/aigatewayclass_types.go`)
   - Defines gateway implementation classes
   - Specifies which controller handles the gateway
   - Enables multiple gateway implementations in same cluster

3. **Validation Webhooks** (`internal/webhook/v1alpha1/`)
   - **AiGateway Webhook**: Validates gateway specs, sets defaults
   - **AiGatewayClass Webhook**: Validates controller references
   - Ensures both `name` and `provider` are set for AI models
   - Validates port ranges (1-65535)

### Project Structure
```
├── api/v1alpha1/           # CRD definitions and types
├── config/                 # Kustomize configurations
│   ├── crd/               # Generated CRD manifests
│   ├── webhook/           # Webhook configurations
│   ├── samples/           # Sample AiGateway resources
│   └── default/           # Default deployment configuration
├── internal/
│   ├── webhook/           # Admission webhook handlers
│   └── controller/        # Empty - implementations in other repos
├── cmd/main.go            # Operator entrypoint (webhooks only)
└── test/e2e/              # End-to-end tests
```

## AiGateway Resource Examples

### Basic Gateway Configuration
```yaml
apiVersion: agentic-layer.ai/v1alpha1
kind: AiGateway
metadata:
  name: my-gateway
spec:
  aiGatewayClassName: litellm
  port: 4000
  aiModels:
    - name: gpt-4
      provider: openai
    - name: claude-3-opus
      provider: anthropic
```

### Multi-Provider Setup
```yaml
apiVersion: agentic-layer.ai/v1alpha1
kind: AiGateway
metadata:
  name: enterprise-gateway
spec:
  aiGatewayClassName: litellm
  port: 8080
  aiModels:
    - name: gpt-4-turbo
      provider: openai
    - name: gpt-4
      provider: azure
    - name: claude-3-sonnet
      provider: anthropic
    - name: gemini-1.5-pro
      provider: gemini
```

### Gateway Class Definition
```yaml
apiVersion: agentic-layer.ai/v1alpha1
kind: AiGatewayClass
metadata:
  name: litellm
spec:
  controller: litellm.agentic-layer.ai/controller
```

## Development Commands

### Building and Testing
```bash
# Generate manifests (CRDs, RBAC, webhooks)
make manifests

# Generate deepcopy code
make generate

# Run unit tests (includes fmt, vet, envtest)
make test

# Run specific test package
go test ./internal/webhook/v1alpha1/... -v

# Lint code
make lint

# Auto-fix linting issues
make lint-fix
```

### Local Development
```bash
# Install CRDs into cluster
make install

# Build docker image
make docker-build

# Load image into Kind cluster
make kind-load

# Deploy operator to cluster
make deploy

# Remove operator from cluster
make undeploy

# Uninstall CRDs
make uninstall
```

### End-to-End Testing
```bash
# Run full e2e suite (creates cluster, runs tests, cleans up)
make test-e2e

# Manual e2e cluster management
make setup-test-e2e                # Create Kind cluster (ai-gateway-operator-test-e2e)
KIND_CLUSTER=ai-gateway-operator-test-e2e go test ./test/e2e/ -v -ginkgo.v
make cleanup-test-e2e              # Delete Kind cluster
```

### Webhook Development
```bash
# Check webhook certificate status
kubectl get certificates -n ai-gateway-operator-system

# Test webhook validation
kubectl apply -f config/samples/v1alpha1_aigateway.yaml --dry-run=server
```

## Important Implementation Details

### AiModel Structure
The `AiModel` type uses **separate fields** for name and provider:
```go
type AiModel struct {
    Name     string `json:"name"`     // e.g., "gpt-4"
    Provider string `json:"provider"`  // e.g., "openai"
}
```

**Previous Format (deprecated)**: `name: openai/gpt-4`
**Current Format**: `name: gpt-4`, `provider: openai`

### Webhook Implementation
- **Defaulting**: Sets default port (4000) if not specified
- **Validation**: Ensures both `name` and `provider` are non-empty for all AI models
- **Port Validation**: Ensures port is in valid range (1-65535)
- **No Controller Logic**: Webhooks only validate/default, no reconciliation

### No Controllers in This Operator
- `internal/controller/` directory exists but is **empty**
- This operator provides only CRDs and webhooks
- Implementation operators (like ai-gateway-litellm-operator) provide reconciliation logic
- Multiple implementations can coexist via different AiGatewayClass values

### Version Management
- Version derived from git tags: `VERSION ?= $(shell git describe --tags --always | sed 's/^v//')`
- Images tagged as: `ghcr.io/agentic-layer/ai-gateway-operator:$(VERSION)`

## Code Architecture Notes

### Separation of Concerns
- **This Operator**: Defines API types, validates inputs via webhooks
- **Implementation Operators**: Watch these CRDs, create actual gateway deployments
- **Pattern**: Similar to Kubernetes Gateway API (defines types, not implementations)

### Webhook vs Controller Responsibilities
- **Webhook**: Validates resource specifications at admission time
- **Controller** (in impl operators): Creates/updates Kubernetes resources based on validated specs
- Webhooks reject invalid resources before they're stored in etcd

### Environment Variables
- `ENABLE_WEBHOOKS=false`: Disable webhooks (useful for local development)

## Common Workflows

### Adding a New Field to AiGateway
1. Update `api/v1alpha1/aigateway_types.go` with new field
2. Add validation logic in `internal/webhook/v1alpha1/aigateway_webhook.go`
3. Add tests in `internal/webhook/v1alpha1/aigateway_webhook_test.go`
4. Run `make manifests generate` to regenerate CRDs and deepcopy code
5. Update sample in `config/samples/v1alpha1_aigateway.yaml`
6. Run `make test` to verify

### Creating New API Resources
Use operator-sdk CLI (preferred method):
```bash
# Create new API
operator-sdk create api --group gateway --version v1alpha1 --kind MyKind

# Create webhook
operator-sdk create webhook --group gateway --version v1alpha1 --kind MyKind --defaulting --programmatic-validation
```

### Testing Webhook Changes
```bash
# Run webhook tests specifically
go test ./internal/webhook/v1alpha1/... -v

# Deploy and test validation
make install deploy
kubectl apply -f config/samples/v1alpha1_aigateway.yaml --dry-run=server
```

## Development Tips

1. **No Controller Development**: This repo contains no controllers; see implementation repos like ai-gateway-litellm-operator
2. **CRD Updates**: After changing API types, always run `make manifests generate`
3. **Webhook Testing**: Use Ginkgo/Gomega tests with envtest for comprehensive coverage
4. **Sample Files**: Keep samples in `config/samples/` up to date with API changes
5. **Documentation**: Update AsciiDoc files in `docs/modules/` for user-facing changes
