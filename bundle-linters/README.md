# odhlint-bundle: OLM Bundle Linter

A static analysis tool for validating Operator Lifecycle Manager (OLM) operator bundles against best practices and requirements.

## Overview

`odhlint-bundle` validates operator bundle structure, ClusterServiceVersion manifests, CRDs, and other Kubernetes resources to catch issues before bundle publication. It implements validation rules derived from official OLM documentation and real-world operator development experience.

## Key Features

- **8 Validation Rules** covering critical OLM requirements and best practices
- **YAML-aware** parsing of Kubernetes manifests
- **Clear categorization** of issues (OLM Requirements, Best Practices, Security, Upgrade)
- **Severity levels** (Error, Warning, Info) with appropriate exit codes
- **Selective rule execution** via `--enable` and `--disable` flags
- **Human-friendly output** with emojis and detailed descriptions

## Installation

### Build from Source

```bash
cd bundle-linters
go build -o odhlint-bundle ./cmd/odhlint-bundle
```

The binary will be created in the `bundle-linters` directory.

### Using Make (from project root)

```bash
cd odh-linter
make build-bundle-linter
```

## Usage

### Basic Validation

```bash
odhlint-bundle ./path/to/bundle/
```

### List All Rules

```bash
odhlint-bundle --list-rules
```

### Selective Rule Execution

```bash
# Run only specific rules
odhlint-bundle --enable ODH-OLM-001,ODH-OLM-002 ./bundle/

# Disable specific rules
odhlint-bundle --disable ODH-OLM-007 ./bundle/
```

### Options

- `--list-rules`: List all available validation rules with descriptions
- `--enable <rule-ids>`: Comma-separated list of rule IDs to enable (default: all)
- `--disable <rule-ids>`: Comma-separated list of rule IDs to disable
- `--no-warnings`: Treat warnings as passing (exit code 0)
- `--version`: Show version information

## Validation Rules

### OLM Requirements (Severity: Error)

#### ODH-OLM-002: Webhook Intercepting Operator Resources

**Critical**: OLM will fail the CSV if webhooks intercept:
- All API groups (`apiGroups: ["*"]`)
- The `operators.coreos.com` group
- `ValidatingWebhookConfigurations` or `MutatingWebhookConfigurations` resources

**Why**: Prevents operators from breaking OLM's ability to manage other operators.

**Example**:
```yaml
# BAD - will cause OLM failure
webhookdefinitions:
- type: ValidatingAdmissionWebhook
  rules:
  - apiGroups: ["operators.coreos.com"]  # FORBIDDEN
    resources: ["*"]

# GOOD
webhookdefinitions:
- type: ValidatingAdmissionWebhook
  rules:
  - apiGroups: ["myapp.example.com"]
    resources: ["myresources"]
```

---

#### ODH-OLM-003: Conversion Webhook Requires AllNamespaces

**Critical**: CSVs with conversion webhooks must support `AllNamespaces` install mode.

**Why**: Conversion webhooks need cluster-wide access to function properly.

**Example**:
```yaml
# BAD
webhookdefinitions:
- type: ConversionWebhook
  ...
installModes:
- type: AllNamespaces
  supported: false  # Must be true!

# GOOD
webhookdefinitions:
- type: ConversionWebhook
  ...
installModes:
- type: AllNamespaces
  supported: true
```

**Fixable**: Yes

---

#### ODH-OLM-010: Conversion Webhook PreserveUnknownFields

**Critical**: CRDs targeted by conversion webhooks must have `spec.preserveUnknownFields: false` or `nil`.

**Why**: Required for proper conversion webhook functionality.

**Example**:
```yaml
# BAD
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  preserveUnknownFields: true  # FORBIDDEN with conversion webhooks

# GOOD
spec:
  preserveUnknownFields: false
```

**Fixable**: Yes

---

### Security Issues (Severity: Error)

#### ODH-OLM-006: PriorityClass globalDefault=true

**Critical**: PriorityClass in bundles must not set `globalDefault: true`.

**Why**: Affects ALL pods cluster-wide without explicit priority, can disrupt unrelated workloads.

**Example**:
```yaml
# BAD
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 1000
globalDefault: true  # FORBIDDEN

# GOOD
globalDefault: false
```

**Fixable**: Yes

---

### Upgrade Issues (Severity: Error)

#### ODH-OLM-004: PDB maxUnavailable=0

**Critical**: PodDisruptionBudget cannot have `maxUnavailable: 0` or `0%`.

**Why**: Prevents node drains, blocks cluster and operator upgrades.

**Example**:
```yaml
# BAD
apiVersion: policy/v1
kind: PodDisruptionBudget
spec:
  maxUnavailable: 0  # FORBIDDEN

# GOOD
spec:
  maxUnavailable: 1
```

---

#### ODH-OLM-005: PDB minAvailable=100%

**Critical**: PodDisruptionBudget cannot have `minAvailable: 100%`.

**Why**: Prevents node drains, blocks cluster and operator upgrades.

**Example**:
```yaml
# BAD
apiVersion: policy/v1
kind: PodDisruptionBudget
spec:
  minAvailable: 100%  # FORBIDDEN

# GOOD
spec:
  minAvailable: 90%
```

---

### Best Practices (Severity: Warning)

#### ODH-OLM-001: Missing minKubeVersion

Operators should specify `spec.minKubeVersion` in their CSV.

**Why**: Prevents installation on incompatible Kubernetes versions, better UX.

**Example**:
```yaml
# RECOMMENDED
spec:
  minKubeVersion: 1.25.0
```

---

#### ODH-OLM-007: Channel Naming Convention

Channels should use recommended prefixes: `stable`, `fast`, `candidate`, `preview`, `alpha`, `beta`.

**Why**: Provides consistent user experience across operators.

**Example**:
```yaml
# DISCOURAGED
annotations:
  operators.operatorframework.io.bundle.channels.v1: myapp-v2

# RECOMMENDED
annotations:
  operators.operatorframework.io.bundle.channels.v1: stable-v2,fast-v2
```

---

## Exit Codes

- **0**: All checks passed (or only warnings with `--no-warnings`)
- **1**: Error-level violations found

## Example Output

```
Loading bundle from: ./bundle/
Running 8 validation rule(s)...

Found 2 issue(s):
  - 1 error(s)
  - 1 warning(s)

❌ [ODH-OLM-002] Webhook 'operator.example.com' intercepts the 'operators.coreos.com' API group. OLM will fail the CSV.
   File: bundle/manifests/operator.clusterserviceversion.yaml
   Category: OLM-Requirement
   Webhooks cannot intercept the operators.coreos.com group. This would break OLM's ability to manage operators.

⚠️  [ODH-OLM-001] ClusterServiceVersion is missing spec.minKubeVersion field
   File: bundle/manifests/operator.clusterserviceversion.yaml
   Category: OLM-Best-Practice
   It is recommended to specify the minimum Kubernetes version your operator supports.

❌ Validation failed: 1 error(s), 1 warning(s)
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Validate operator bundle
  run: |
    ./odhlint-bundle ./bundle/
```

### GitLab CI

```yaml
bundle-lint:
  script:
    - ./odhlint-bundle ./bundle/
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit
if [ -d "bundle" ]; then
  odhlint-bundle ./bundle/ || exit 1
fi
```

## Bundle Structure

`odhlint-bundle` expects the standard operator bundle structure:

```
bundle/
├── manifests/
│   ├── operator.clusterserviceversion.yaml
│   ├── crd1.yaml
│   ├── crd2.yaml
│   └── ...other resources
└── metadata/
    └── annotations.yaml
```

## Comparison with operator-sdk validate

`odhlint-bundle` complements `operator-sdk bundle validate`:

- **operator-sdk**: Schema validation, API structure, bundle format
- **odhlint-bundle**: OLM-specific requirements, operational best practices, upgrade safety

**Recommendation**: Run both tools in your CI pipeline.

## Development

### Adding New Rules

1. Create a new file in `pkg/rules/`: `olmXXX_description.go`
2. Implement the `Rule` interface
3. Add the rule to `GetAllRules()` in `pkg/rules/registry.go`
4. Update this README with rule documentation
5. Add test cases

### Rule Interface

```go
type Rule interface {
    ID() string          // e.g., "ODH-OLM-001"
    Name() string        // e.g., "missing-minkubeversion"
    Category() Category  // OLMRequirement, OLMBestPractice, Security, Upgrade
    Severity() Severity  // Error, Warning, Info
    Description() string // Detailed explanation
    Validate(bundle *Bundle) []Violation
    Fixable() bool       // Can be auto-fixed?
}
```

## Provenance

These rules were derived from:
- Official OLM documentation ([olm.operatorframework.io](https://olm.operatorframework.io))
- OpenShift Operator best practices
- Real-world operator development experience
- Analysis of opendatahub-operator and other ODH components

See `../../OLM_LINTER_PATTERNS.md` for detailed pattern analysis and sources.

## Related Tools

- [`odhlint`](../linters/odhlint): Go code linters for ODH operators (AST-based)
- [`operator-sdk bundle validate`](https://sdk.operatorframework.io/docs/cli/operator-sdk_bundle_validate/): Official bundle validation
- [`opm alpha bundle validate`](https://github.com/operator-framework/operator-registry): Registry-based validation

## License

Same as parent odh-linter project.

## Contributing

Contributions welcome! Please:
1. Add test cases for new rules
2. Update documentation
3. Follow existing code patterns
4. Ensure all existing tests pass

