# ODH Linter

Custom linters for OpenDataHub operator projects, extracted from real PR reviews, refactorings, and OLM documentation.

## Tools

ODH Linter provides **two complementary linting tools**:

1. **`odhlint`**: Go code linters (12 rules) - Static analysis of Go source code
2. **`odhlint-bundle`**: OLM bundle linters (8 rules) - Validation of operator bundle manifests

Both tools work together to ensure high-quality operator development.

## Quick Start

### Go Code Linter (`odhlint`)

```bash
# Build from source
git clone https://github.com/opendatahub-io/odh-linter
cd odh-linter
make build

# List all rules
./cmd/odhlint -list-rules

# Run on your Go code
go vet -vettool=$(pwd)/cmd/odhlint ./...
```

### Bundle Linter (`odhlint-bundle`)

```bash
# Build (included in make build)
make build

# List all rules
./bundle-linters/odhlint-bundle --list-rules

# Validate your operator bundle
./bundle-linters/odhlint-bundle ./bundle/
```

## What is ODH Linter?

ODH Linter is a collection of **20 custom linting rules** (12 Go + 8 OLM) specifically designed for OpenDataHub operator development. All rules were extracted from:

- **3,785 PR review comments** from opendatahub-operator repository
- **Refactoring PRs** by senior maintainers (e.g., PR #2769)
- **Official OLM documentation** and best practices

### Why `go vet` is Required

ODH Linter is built using **Go's official static analysis framework** (`golang.org/x/tools/go/analysis`). This framework integrates with Go's standard tooling via the `-vettool` flag.

**How it works**:
1. `go vet` compiles your code and creates a typed AST (Abstract Syntax Tree)
2. It passes this AST to the analyzer specified by `-vettool`
3. The analyzer inspects the AST and reports issues

**Why not a standalone command?**
- Reuses Go's compiler and type checker (faster, more accurate)
- Integrates seamlessly with existing Go tooling
- Works with standard Go build caching
- Compatible with IDEs and CI/CD systems that already use `go vet`

**Alternative invocation methods**:
```bash
# Direct via go vet (recommended)
go vet -vettool=odhlint ./...

# Via golangci-lint (integration)
golangci-lint run --enable=odhlint

# Standalone analysis (advanced)
odhlint ./...  # Also works, but bypasses go vet's optimizations
```

The `go vet` integration ensures ODH Linter works efficiently with large codebases and plays nicely with your existing Go development workflow.

## Rule Categories

### 🔍 Code-Level Checks (7 rules)

Extracted from PR review comments:

| Rule ID | Name | Description | Severity |
|---------|------|-------------|----------|
| ODH-ERR-001 | `doublewrap` | Redundant error wrapping | LOW |
| ODH-SEC-001 | `insecureskipverify` | InsecureSkipVerify disables TLS | HIGH ⚠️ |
| ODH-TEST-001 | `testhelper` | Missing t.Helper() | MEDIUM |
| ODH-WEBHOOK-001 | `admissionwebhook` | Suspicious webhook Denied() | MEDIUM |
| ODH-TYPE-001 | `typeassert` | Type assertion without ok check | HIGH ⚠️ |
| ODH-STYLE-001 | `fmtsprintf` | Unnecessary fmt.Sprintf | LOW |
| ODH-STYLE-002 | `selectcase` | Single-case select | LOW |

### 🏗️ Architectural Checks (5 rules)

Extracted from refactoring PRs:

| Rule ID | Name | Description | Severity |
|---------|------|-------------|----------|
| ODH-ARCH-001 | `clusterconfig` | Direct cluster config access | MEDIUM |
| ODH-ARCH-002 | `depdirection` | Wrong dependency direction | HIGH ⚠️ |
| ODH-ARCH-003 | `dupedef` | Duplicate type definitions | LOW |
| ODH-ARCH-004 | `testlocation` | Tests in wrong package (unit tests only) | LOW |
| ODH-ARCH-005 | `boolforenum` | Boolean return for enum | MEDIUM |

### 📦 OLM Bundle Checks (8 rules)

Extracted from OLM documentation and best practices:

| Rule ID | Name | Description | Severity |
|---------|------|-------------|----------|
| ODH-OLM-001 | `missing-minkubeversion` | Missing spec.minKubeVersion | Warning |
| ODH-OLM-002 | `webhook-intercepts-operators` | Webhook intercepts OLM resources | Error ❌ |
| ODH-OLM-003 | `conversion-webhook-allnamespaces` | Conversion webhook needs AllNamespaces | Error ❌ |
| ODH-OLM-004 | `pdb-maxunavailable-zero` | PDB maxUnavailable=0 blocks drains | Error ❌ |
| ODH-OLM-005 | `pdb-minavailable-hundred` | PDB minAvailable=100% blocks drains | Error ❌ |
| ODH-OLM-006 | `priorityclass-globaldefault` | PriorityClass globalDefault=true | Error ❌ |
| ODH-OLM-007 | `channel-naming` | Non-standard channel naming | Warning |
| ODH-OLM-010 | `conversion-preserveunknownfields` | CRD preserveUnknownFields with conversion | Error ❌ |

See [`bundle-linters/README.md`](bundle-linters/README.md) for detailed documentation.

## Usage

### Run All Linters

```bash
go vet -vettool=odhlint ./...
```

### Run Specific Linters

```bash
# Only HIGH severity checks
go vet -vettool=odhlint \
  -doublewrap=false \
  -fmtsprintf=false \
  -selectcase=false \
  -dupedef=false \
  -testlocation=false \
  -testhelper=false \
  -admissionwebhook=false \
  -boolforenum=false \
  -clusterconfig=false \
  ./...
```

### Disable Specific Linters

```bash
# Disable specific linters if needed
go vet -vettool=odhlint -doublewrap=false -fmtsprintf=false ./...
```

## Important Notes

### testlocation Linter (ODH-ARCH-004)

**Now Fixed**: The `testlocation` linter previously generated many false positives. It has been updated to:

✅ **Skip integration/e2e tests** - These are SUPPOSED to import many packages  
✅ **Ignore utility packages** - Kubernetes APIs, test utilities, and common helpers are OK to import  
✅ **Focus on unit tests** - Only flags unit tests testing business logic from other packages

**What it checks**: Unit tests should be in the same package as the code they test.

**What it ignores**:
- E2E/integration tests (`tests/e2e/`, `tests/integration/`)
- Kubernetes utility imports (`k8s.io/apimachinery/pkg/api/errors`, etc.)
- Test utility imports (`pkg/utils/test/`, `pkg/testutil/`)
- Common utilities (`pkg/utils/`, `pkg/common/`)

See [`TESTLOCATION_FIX.md`](TESTLOCATION_FIX.md) for detailed fix documentation.

### clusterconfig Linter (ODH-ARCH-001)

**Now Fixed**: The `clusterconfig` linter previously flagged test files. It has been updated to:

✅ **Skip all test files** (`*_test.go`) - Tests need to create fixtures directly  
✅ **Focus on production code** - Only flags production code using cluster config types  
✅ **Allow test setup** - Unit and E2E tests can use `configv1.Authentication`, `configv1.ClusterVersion`, etc.

**What it checks**: Production code should use `pkg/cluster/` abstractions instead of direct OpenShift API types.

**What it ignores**:
- All test files (`*_test.go`) - unit tests and E2E tests
- E2E/integration test directories
- Code in `pkg/cluster/` itself

See [`CLUSTERCONFIG_FIX.md`](CLUSTERCONFIG_FIX.md) and [`E2E_TEST_LINTING.md`](E2E_TEST_LINTING.md) for detailed documentation.

### typeassert Linter (ODH-TYPE-001)

**Now Supports `//nolint` Comments**: The `typeassert` linter respects developer decisions:

✅ **Recognizes `//nolint:typeassert`** - Suppress warnings for assertions you've verified are safe  
✅ **Also recognizes `//nolint:forcetypeassert`** - Common golangci-lint directive  
✅ **Trusts informed decisions** - When developers know the type is guaranteed

**When to use `//nolint`**:
- Kubernetes API guarantees (e.g., container names are always strings)
- After exhaustive validation elsewhere
- In generated code where types are guaranteed

**When NOT to use**:
- User input or external data
- When you can easily add `, ok` check
- "It works in my tests" (production data may differ)

See [`NOLINT_SUPPORT.md`](NOLINT_SUPPORT.md) for detailed guidelines and examples.

## Example Output

```
internal/controller/services/auth/auth.go:24:24: 
  clusterconfig: direct instantiation of cluster config type configv1.Authentication 
  outside pkg/cluster/; use cluster package abstractions instead

api/dscinitialization/v1/conversion.go:26:9: 
  typeassert: type assertion without ok check can panic at runtime; 
  use comma-ok idiom: x, ok := y.(*dsciv2.DSCInitialization)
```

## Real Results

When run on `opendatahub-operator`, ODH Linter found:

- **19 unsafe type assertions** (ODH-TYPE-001) - can cause runtime panics
- **15 direct cluster config accesses** (ODH-ARCH-001) - architectural violations
- **5 wrong dependency directions** (ODH-ARCH-002) - pkg/ importing internal/
- **2 InsecureSkipVerify** issues (ODH-SEC-001) - security vulnerabilities
- **1 redundant error wrapping** (ODH-ERR-001)

**Total: 42+ real issues found** that slipped through code review.

## Documentation

- [`linters/odhlint/README.md`](linters/odhlint/README.md) - Detailed usage guide
- [`linters/RULE_ID_MAPPING.md`](linters/RULE_ID_MAPPING.md) - Quick reference
- [`PROVENANCE.md`](PROVENANCE.md) - How rules were extracted

## Individual Linters

Each linter can also be run standalone:

```bash
# Build individual linter
cd linters/typeassert
go build -o typeassert ./cmd/typeassert

# Run standalone
go vet -vettool=./typeassert ./...
```

## Installation

### From Source

```bash
git clone https://github.com/opendatahub-io/odh-linter
cd odh-linter/linters/odhlint
go build -o odhlint ./cmd/odhlint
sudo mv odhlint /usr/local/bin/
```

### Using go install

```bash
go install github.com/opendatahub-io/odh-linter/linters/odhlint/cmd/odhlint@latest
```

### Docker

```bash
docker run --rm -v $(pwd):/src ghcr.io/opendatahub-io/odh-linter:latest ./...
```

## CI/CD Integration

### GitHub Actions

```yaml
name: ODH Lint
on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install ODH Linter
        run: go install github.com/opendatahub-io/odh-linter/linters/odhlint/cmd/odhlint@latest
      
      - name: Run ODH Linter
        run: go vet -vettool=$(which odhlint) ./...
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

if command -v odhlint &> /dev/null; then
    echo "Running ODH Linter..."
    go vet -vettool=odhlint ./...
    if [ $? -ne 0 ]; then
        echo "❌ ODH Linter found issues. Fix them before committing."
        exit 1
    fi
fi
```

### golangci-lint Integration

Add to `.golangci.yml`:

```yaml
linters-settings:
  govet:
    enable-all: true
    settings:
      odhlint:
        path: ./tools/odhlint
```

## Architecture

```
odh-linter/
├── linters/
│   ├── admissionwebhook/    # ODH-WEBHOOK-001
│   ├── boolforenum/         # ODH-ARCH-005
│   ├── clusterconfig/       # ODH-ARCH-001
│   ├── depdirection/        # ODH-ARCH-002
│   ├── doublewrap/          # ODH-ERR-001
│   ├── dupedef/             # ODH-ARCH-003
│   ├── fmtsprintf/          # ODH-STYLE-001
│   ├── insecureskipverify/  # ODH-SEC-001
│   ├── selectcase/          # ODH-STYLE-002
│   ├── testhelper/          # ODH-TEST-001
│   ├── testlocation/        # ODH-ARCH-004
│   ├── typeassert/          # ODH-TYPE-001
│   └── odhlint/             # Unified tool
├── README.md                # This file
└── PROVENANCE.md            # How rules were extracted
```

## Provenance

All linters were extracted using a hybrid human-AI workflow:

1. **Data Collection**: Scraped 3,785 PR review comments from opendatahub-operator
2. **Pattern Analysis**: Python scripts + multi-signal importance scoring
3. **Human Review**: Manual identification of lintable patterns
4. **Linter Generation**: AI-assisted (Claude) Go AST code generation
5. **Validation**: Tested on real codebase, found 42+ issues

### Key Contributors

- **Data Source**: opendatahub-io/opendatahub-operator repository
- **Refactoring Reference**: PR #2769 by @zdtsw
- **Review Comments**: 15+ reviewers including zdtsw, bartoszmajsak, ykaliuta
- **Tool Development**: Hybrid Python + AI (Claude via Cursor)

## Development

### Adding New Linters

1. Create new linter directory:
   ```bash
   mkdir -p linters/newlinter/cmd/newlinter
   ```

2. Implement analyzer:
   ```go
   // linters/newlinter/newlinter.go
   package newlinter
   
   var Analyzer = &analysis.Analyzer{
       Name: "newlinter",
       Doc:  "Description",
       Run:  run,
   }
   ```

3. Add to odhlint:
   ```go
   // linters/odhlint/cmd/odhlint/main.go
   import "github.com/opendatahub-io/odh-linter/linters/newlinter"
   
   analyzers := []*analysis.Analyzer{
       withRuleID(newlinter.Analyzer, "ODH-XXX-###", "Description"),
   }
   ```

### Testing

```bash
# Test individual linter
cd linters/typeassert
go test ./...

# Test unified tool
cd linters/odhlint
go build -o odhlint ./cmd/odhlint
./odhlint -list-rules
```

## Statistics

- **Total Linters**: 12
- **Lines of Code**: ~2,000 (analyzer code)
- **Test Cases**: 50+ test scenarios
- **Real Issues Found**: 42+ in opendatahub-operator
- **False Positive Rate**: <10%
- **Development Cost**: ~$10 (AI-assisted)
- **Development Time**: ~20 hours

## License

Apache 2.0

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new linters
4. Update documentation
5. Submit a pull request

## Support

- **Issues**: https://github.com/opendatahub-io/odh-linter/issues
- **Discussions**: https://github.com/opendatahub-io/odh-linter/discussions
- **Documentation**: https://github.com/opendatahub-io/odh-linter/tree/main/linters/odhlint

## Acknowledgments

This project was inspired by the need to capture institutional knowledge from PR reviews and make it reusable through static analysis. Special thanks to:

- OpenDataHub operator maintainers and reviewers
- The Go static analysis framework authors
- Claude AI for code generation assistance

---

**Built with ❤️ for the OpenDataHub community**

