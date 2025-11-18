# ODH Linter Quick Start

Get started with ODH Linter in 5 minutes.

## Installation

### Option 1: Build from Source (Recommended)

```bash
# Clone the repository
cd /path/to/your/projects
git clone https://github.com/opendatahub-io/odh-linter
cd odh-linter

# Build using Makefile
make build

# Verify it works
./linters/odhlint/odhlint -list-rules

# Move to PATH (optional)
sudo cp linters/odhlint/odhlint /usr/local/bin/
```

**Alternative**: Build manually without Makefile:
```bash
cd odh-linter/linters/odhlint
go build -o odhlint ./cmd/odhlint
```

**Note**: The project structure is:
```
odh-linter/
├── Makefile            ← Simple build script
└── linters/odhlint/
    ├── cmd/
    │   └── odhlint/    ← main.go is here
    │       └── main.go
    ├── odhlint.go      ← shared analyzer code
    └── go.mod
```

### Option 2: Go Install

```bash
go install github.com/opendatahub-io/odh-linter/linters/odhlint/cmd/odhlint@latest
```

## Verify Installation

```bash
odhlint -list-rules
```

You should see output like:

```
OpenDataHub Custom Linter Rules
================================

Error Handling:
  ODH-ERR-001     doublewrap                Redundant error wrapping
...
```

## Run on Your Project

### Basic Usage

```bash
# Go to your Go project directory
cd /path/to/your/opendatahub-project

# Run all linters
go vet -vettool=$(which odhlint) ./...
```

**Why use `go vet`?** ODH Linter uses Go's standard analysis framework, which integrates via `go vet -vettool`. This gives you:
- Faster analysis (reuses Go's compiler and type checker)
- Better accuracy (full type information available)
- Standard Go tooling integration
- IDE and CI/CD compatibility

The analyzer can also run standalone (`odhlint ./...`), but `go vet` is recommended for performance.

### Disable False Positives

```bash
# testlocation has known false positives with external deps
go vet -vettool=$(which odhlint) -testlocation=false ./...
```

### Run on Specific Package

```bash
go vet -vettool=$(which odhlint) ./internal/controller/...
```

## Interpreting Results

### Example Output

```
internal/controller/services/auth/auth.go:24:24: 
  clusterconfig: direct instantiation of cluster config type configv1.Authentication 
  outside pkg/cluster/; use cluster package abstractions instead
```

Breaking it down:
- `internal/controller/services/auth/auth.go:24:24` - File and position
- `clusterconfig` - Analyzer name (maps to ODH-ARCH-001)
- Rest - Description of the issue

### Look Up Rule ID

```bash
# See the analyzer → rule ID mapping
cat linters/RULE_ID_MAPPING.md | grep clusterconfig
```

Result:
```
| `clusterconfig` | ODH-ARCH-001 | Architecture | MEDIUM |
```

## Common Scenarios

### Scenario 1: Pre-commit Check

```bash
#!/bin/bash
# .git/hooks/pre-commit

go vet -vettool=$(which odhlint) -testlocation=false ./...
if [ $? -ne 0 ]; then
    echo "❌ ODH Linter found issues"
    exit 1
fi
```

### Scenario 2: CI/CD

```yaml
# .github/workflows/lint.yml
- name: Run ODH Linter
  run: |
    go install github.com/opendatahub-io/odh-linter/linters/odhlint/cmd/odhlint@latest
    go vet -vettool=$(which odhlint) -testlocation=false ./...
```

### Scenario 3: Only HIGH Severity

```bash
# Run only HIGH severity checks (security, panics, architecture)
go vet -vettool=odhlint \
  -doublewrap=false \
  -fmtsprintf=false \
  -selectcase=false \
  -testhelper=false \
  -admissionwebhook=false \
  -dupedef=false \
  -testlocation=false \
  -boolforenum=false \
  -clusterconfig=false \
  ./...

# This runs: insecureskipverify, typeassert, depdirection
```

## What to Expect

Running ODH Linter on opendatahub-operator found:

- **19 type assertion bugs** (can panic)
- **15 architectural violations** (direct cluster config access)
- **5 dependency direction issues** (pkg/ importing internal/)
- **2 security issues** (InsecureSkipVerify)
- **1 code quality issue** (redundant error wrapping)

Total: **42+ real issues** that made it past human review.

## Understanding Severity

### 🔴 HIGH - Fix Immediately

- **ODH-SEC-001** (insecureskipverify) - Security vulnerability
- **ODH-TYPE-001** (typeassert) - Can panic at runtime
- **ODH-ARCH-002** (depdirection) - Wrong dependency direction

### 🟡 MEDIUM - Should Fix

- **ODH-TEST-001** (testhelper) - Test quality
- **ODH-WEBHOOK-001** (admissionwebhook) - Logic errors
- **ODH-ARCH-001** (clusterconfig) - Architecture violations
- **ODH-ARCH-005** (boolforenum) - API design

### 🟢 LOW - Nice to Fix

- **ODH-ERR-001** (doublewrap) - Code clarity
- **ODH-STYLE-001**, **ODH-STYLE-002** - Style
- **ODH-ARCH-003**, **ODH-ARCH-004** - Organization

## Troubleshooting

### "Command not found: odhlint"

```bash
# Make sure it's built
cd odh-linter/linters/odhlint
go build -o odhlint ./cmd/odhlint

# Add to PATH or use full path
export PATH=$PATH:$(pwd)
# or
go vet -vettool=$(pwd)/odhlint ./...
```

### Too Many False Positives

```bash
# Disable testlocation (known issue with external deps)
go vet -vettool=odhlint -testlocation=false ./...

# Or disable specific analyzers
go vet -vettool=odhlint \
  -testlocation=false \
  -dupedef=false \
  ./...
```

### "internal error: package without types"

This is a Go tooling issue. Try:

```bash
# Clean and rebuild
go clean -cache
go mod tidy

# Run on specific packages instead of ./...
go vet -vettool=odhlint ./pkg/... ./internal/...
```

## Next Steps

1. **Read the full documentation**: [`linters/odhlint/README.md`](linters/odhlint/README.md)
2. **Check rule mappings**: [`linters/RULE_ID_MAPPING.md`](linters/RULE_ID_MAPPING.md)
3. **Understand provenance**: [`PROVENANCE.md`](PROVENANCE.md)
4. **Integrate in CI/CD**: See README for examples

## Get Help

- **Issues**: https://github.com/opendatahub-io/odh-linter/issues
- **Discussions**: https://github.com/opendatahub-io/odh-linter/discussions
- **Documentation**: https://github.com/opendatahub-io/odh-linter

## Summary

```bash
# Three commands to get started:
git clone https://github.com/opendatahub-io/odh-linter
cd odh-linter/linters/odhlint && go build -o odhlint ./cmd/odhlint
go vet -vettool=$(pwd)/odhlint -testlocation=false /path/to/your/project/...
```

That's it! 🎉

