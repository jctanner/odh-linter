# errordemote - Error Demotion Linter

**Rule ID**: `ODH-ARCH-006`

Detects the "fail-fast vs resilient" pattern where errors are caught but only logged instead of being returned to the caller.

## Problem

In production Kubernetes operators, error handling philosophy matters:

```go
// Pattern: Error is caught but only logged
if value, err := getConfig(ctx, cli); err == nil {
    config.Value = value
} else {
    log.Info("couldn't get config", "error", err)  // ⚠️ Error hidden
}
// Code continues with zero value
```

This pattern can:
- Hide critical failures
- Make debugging difficult
- Violate fail-fast principles
- Create inconsistent behavior

## Rule

The linter flags cases where:
1. A function returns `(value, error)`
2. Error is caught in an if statement
3. Error branch **only logs** (doesn't return)
4. Log level is Info/Debug/Warn (not Error)
5. Code continues with a default value

## Background

This pattern was identified in PR [#1898](https://github.com/opendatahub-io/opendatahub-operator/pull/1898) during a debate about FIPS detection:

**CodeRabbit suggested** (resilient):
```go
if fipsEnabled, err := IsFipsEnabled(ctx, cli); err == nil {
    c.FipsEnabled = fipsEnabled
} else {
    logf.FromContext(ctx).Info("could not determine FIPS status, defaulting to false", "error", err)
}
```

**zdtsw questioned**: 
> "if we fail to get value from configmap, we default bool to false for FipsEnabled, but no error only info ?"

Both viewpoints are valid - the key is **making the decision explicit**.

## Solutions

### Option 1: Add //nolint with Justification

```go
//nolint:errordemote // ConfigMap may not exist on non-OCP clusters; safe to default to false
if value, err := getConfig(ctx, cli); err == nil {
    config.Value = value
} else {
    log.Info("couldn't get config", "error", err)
}
```

### Option 2: Document Resilience Decision

```go
// RESILIENCE: FIPS detection is non-critical; safe to continue with false default
if fipsEnabled, err := IsFipsEnabled(ctx, cli); err == nil {
    c.FipsEnabled = fipsEnabled
} else {
    log.Info("FIPS detection failed", "error", err)
}
```

### Option 3: Return the Error (Fail-Fast)

```go
// For critical configuration
value, err := getAPIKey(ctx, cli)
if err != nil {
    return fmt.Errorf("failed to get API key: %w", err)
}
config.APIKey = value
```

### Option 4: Use Error Log Level

```go
// Log at Error level to make it prominent
if value, err := getConfig(ctx, cli); err == nil {
    config.Value = value
} else {
    log.Error(err, "couldn't get config")  // Error level = more visible
}
```

## When to Fail-Fast vs Be Resilient

| Scenario | Recommendation | Why |
|----------|----------------|-----|
| **API credentials** | Fail-fast | Silent failure = security risk |
| **Database connection** | Fail-fast | Can't function without it |
| **Required CRDs** | Fail-fast | Operator can't work |
| **Feature flags** | Resilient | Optional functionality |
| **Telemetry config** | Resilient | Observability not critical |
| **Performance tuning** | Resilient | Safe defaults exist |
| **FIPS detection** | Resilient | Environment-specific, has safe default |

## Examples

### ❌ Flagged: Critical Config Silently Ignored

```go
// No justification for hiding this error
if apiKey, err := getAPIKey(ctx); err == nil {
    config.APIKey = apiKey
} else {
    log.Info("couldn't get API key", "error", err)  // 🚨 Critical failure hidden!
}
```

### ✅ OK: Documented Resilience

```go
// RESILIENCE: FIPS is optional; operator functions normally with false default
if fips, err := detectFIPS(ctx); err == nil {
    config.FIPS = fips
} else {
    log.Info("FIPS detection failed, assuming non-FIPS cluster", "error", err)
}
```

### ✅ OK: Returns Error

```go
creds, err := getCredentials(ctx)
if err != nil {
    return fmt.Errorf("cannot start without credentials: %w", err)
}
```

### ✅ OK: Suppressed with nolint

```go
//nolint:errordemote // Optional telemetry; operator works fine without it
if endpoint, err := getTelemetryEndpoint(); err == nil {
    config.Telemetry = endpoint
} else {
    log.Debug("telemetry not configured", "error", err)
}
```

## Keywords for Automatic Suppression

The linter automatically skips cases with these comment keywords:
- `RESILIENCE:` or `RESILIENT:`
- `non-critical`
- `optional`
- `safe to ignore`
- `safe to continue`
- `safe default`
- `may not exist`

Example:
```go
// This ConfigMap may not exist on older cluster versions; safe to continue
if cm, err := getCM(ctx); err == nil {
    processConfigMap(cm)
} else {
    log.Info("ConfigMap not found", "error", err)
}
```

## Usage

### Standalone

```bash
go install github.com/opendatahub-io/odh-linter/linters/errordemote/cmd/errordemote@latest
go vet -vettool=$(which errordemote) ./...
```

### With odhlint

```bash
odhlint ./...
# or
go vet -vettool=./odhlint ./...
```

## Integration with CI/CD

```yaml
- name: Check error handling
  run: |
    go vet -vettool=./odhlint ./... 2>&1 | grep "ODH-ARCH-006" || true
```

## Related Rules

- **ODH-ERR-001** (`doublewrap`) - Redundant error wrapping
- **ODH-ARCH-002** (`depdirection`) - Dependency direction violations

## Philosophy

This linter doesn't enforce a single philosophy. Instead, it requires **explicit documentation** of error handling decisions. Teams can decide their own policies, but those decisions must be visible in the code.

The goal: Make implicit architectural decisions explicit! 🎯

