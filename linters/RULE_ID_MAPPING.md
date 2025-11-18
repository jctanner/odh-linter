# ODH Linter Rule ID Mapping

Quick reference for mapping analyzer names in output to rule IDs.

## Quick Lookup

When you see an error like:
```
file.go:10:5: clusterconfig: direct access to cluster config type...
```

The analyzer name is `clusterconfig`, which maps to rule ID **ODH-ARCH-001**.

---

## Complete Mapping

| Analyzer Name | Rule ID | Category | Severity |
|---------------|---------|----------|----------|
| `doublewrap` | ODH-ERR-001 | Error Handling | LOW |
| `insecureskipverify` | ODH-SEC-001 | Security | HIGH ⚠️ |
| `testhelper` | ODH-TEST-001 | Testing | MEDIUM |
| `admissionwebhook` | ODH-WEBHOOK-001 | Webhooks | MEDIUM |
| `typeassert` | ODH-TYPE-001 | Type Safety | HIGH ⚠️ |
| `fmtsprintf` | ODH-STYLE-001 | Code Style | LOW |
| `selectcase` | ODH-STYLE-002 | Code Style | LOW |
| `clusterconfig` | ODH-ARCH-001 | Architecture | MEDIUM |
| `depdirection` | ODH-ARCH-002 | Architecture | HIGH ⚠️ |
| `dupedef` | ODH-ARCH-003 | Architecture | LOW |
| `testlocation` | ODH-ARCH-004 | Architecture | LOW |
| `boolforenum` | ODH-ARCH-005 | Architecture | MEDIUM |
| `errordemote` | ODH-ARCH-006 | Architecture | MEDIUM |

---

## By Rule ID

| Rule ID | Analyzer | One-Line Description |
|---------|----------|---------------------|
| ODH-ERR-001 | `doublewrap` | Redundant error wrapping |
| ODH-SEC-001 | `insecureskipverify` | TLS verification disabled |
| ODH-TEST-001 | `testhelper` | Missing t.Helper() |
| ODH-WEBHOOK-001 | `admissionwebhook` | Negation in webhook Denied() |
| ODH-TYPE-001 | `typeassert` | Type assertion without ok |
| ODH-STYLE-001 | `fmtsprintf` | Unnecessary fmt.Sprintf |
| ODH-STYLE-002 | `selectcase` | Single-case select |
| ODH-ARCH-001 | `clusterconfig` | Direct cluster config access |
| ODH-ARCH-002 | `depdirection` | pkg/ imports internal/ |
| ODH-ARCH-003 | `dupedef` | Duplicate type definition |
| ODH-ARCH-004 | `testlocation` | Test in wrong package |
| ODH-ARCH-005 | `boolforenum` | Bool return for enum |
| ODH-ARCH-006 | `errordemote` | Error logged instead of returned |

---

## By Category

### 🔴 HIGH Severity (Fix Immediately)

- **ODH-SEC-001** (`insecureskipverify`) - Security vulnerability
- **ODH-TYPE-001** (`typeassert`) - Can cause runtime panics
- **ODH-ARCH-002** (`depdirection`) - Architectural violation

### 🟡 MEDIUM Severity (Should Fix)

- **ODH-TEST-001** (`testhelper`) - Test quality issue
- **ODH-WEBHOOK-001** (`admissionwebhook`) - Logic error risk
- **ODH-ARCH-001** (`clusterconfig`) - Design pattern violation
- **ODH-ARCH-005** (`boolforenum`) - API design issue
- **ODH-ARCH-006** (`errordemote`) - Error handling policy violation

### 🟢 LOW Severity (Nice to Fix)

- **ODH-ERR-001** (`doublewrap`) - Code clarity
- **ODH-STYLE-001** (`fmtsprintf`) - Minor optimization
- **ODH-STYLE-002** (`selectcase`) - Code simplification
- **ODH-ARCH-003** (`dupedef`) - Code organization
- **ODH-ARCH-004** (`testlocation`) - Test organization

---

## Suppression

### Disable Specific Rule

```bash
# Disable by analyzer name
go vet -vettool=odhlint -testlocation=false ./...
```

### Disable in Code

```go
//lint:ignore typeassert Intentional panic for test assertion
user := obj.(*User)
```

Or use build tags:

```go
//go:build !odhlint
// +build !odhlint

// This file is not checked by odhlint
```

---

## CI/CD Integration

### GitHub Actions

```yaml
- name: ODH Lint
  run: |
    go vet -vettool=./odhlint ./... || true
    # Parse output and create annotations with rule IDs
```

### Fail on HIGH Severity Only

```bash
# Run all, but only fail on HIGH severity rules
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

# This runs only: insecureskipverify, typeassert, depdirection
```

---

## Rule ID Format

```
ODH-<CATEGORY>-<NUMBER>

ODH         = OpenDataHub
CATEGORY    = ERR|SEC|TEST|WEBHOOK|TYPE|STYLE|ARCH
NUMBER      = 001, 002, 003, ...
```

### Categories

- **ERR**: Error handling
- **SEC**: Security
- **TEST**: Testing practices
- **WEBHOOK**: Admission webhooks
- **TYPE**: Type safety
- **STYLE**: Code style
- **ARCH**: Architecture/design

---

## Examples with Rule IDs

### Example 1: Type Assertion

**Output**:
```
api/v1/conversion.go:26:9: typeassert: type assertion without ok check can panic
```

- **Analyzer**: `typeassert`
- **Rule ID**: ODH-TYPE-001
- **Severity**: HIGH
- **Action**: Add comma-ok check

### Example 2: Cluster Config Access

**Output**:
```
internal/controller/services/auth/auth.go:24:24: clusterconfig: direct access to cluster config type
```

- **Analyzer**: `clusterconfig`
- **Rule ID**: ODH-ARCH-001  
- **Severity**: MEDIUM
- **Action**: Use pkg/cluster/ abstraction

### Example 3: Wrong Dependency

**Output**:
```
pkg/controller/reconciler/reconciler.go:25:2: depdirection: pkg/ should not import internal/
```

- **Analyzer**: `depdirection`
- **Rule ID**: ODH-ARCH-002
- **Severity**: HIGH
- **Action**: Refactor package structure

---

## Get Full Rule Details

```bash
# List all rules
odhlint -list-rules

# Get analyzer help
odhlint -h

# Get specific analyzer docs
go doc github.com/opendatahub-io/rule-creator/linters/typeassert
```

---

## Quick Commands

```bash
# Show this mapping
odhlint -list-rules

# Count issues by analyzer
go vet -vettool=odhlint ./... 2>&1 | grep -oP '^\S+:\d+:\d+: \K\w+' | sort | uniq -c | sort -rn

# Find all HIGH severity issues
go vet -vettool=odhlint ./... 2>&1 | grep -E 'typeassert|insecureskipverify|depdirection'

# Export results to file
go vet -vettool=odhlint ./... 2>&1 | tee odh-lint-results.txt
```

---

## Provenance

All rule IDs trace back to actual OpenDataHub operator development:

- **Code Rules** (ODH-ERR, ODH-SEC, etc.): Extracted from 3,785 PR review comments
- **Arch Rules** (ODH-ARCH): Extracted from refactoring PR #2769 by zdtsw

See `COMPLETE_LINTER_SUITE.md` for full provenance details.

