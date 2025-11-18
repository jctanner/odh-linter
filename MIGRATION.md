# Migration from rule-creator to odh-linter

This document explains the restructuring of the linter project into a standalone package.

## What Changed

### Before (rule-creator/linters/)

```
2025_11_10_rhoai_merge_stats/
├── rule-creator/
│   ├── linters/
│   │   ├── doublewrap/
│   │   ├── typeassert/
│   │   ├── odhlint/
│   │   └── ... (10 more)
│   ├── src/
│   │   ├── loaders/
│   │   └── processors/
│   └── data/
└── go-github-scraper/
```

### After (odh-linter/)

```
2025_11_10_rhoai_merge_stats/
├── odh-linter/                    # NEW: Standalone project
│   ├── linters/
│   │   ├── doublewrap/
│   │   ├── typeassert/
│   │   ├── odhlint/
│   │   └── ... (10 more)
│   ├── README.md                  # Main documentation
│   ├── PROVENANCE.md              # Rule origins
│   ├── QUICK_START.md             # Getting started guide
│   └── .gitignore
├── rule-creator/                  # Research & analysis project
│   ├── src/                       # Python analysis tools
│   ├── data/                      # Processed data
│   └── docs/                      # Research documentation
└── go-github-scraper/             # Data collection tool
```

## Why the Move?

### 1. Separation of Concerns

- **rule-creator**: Research project for mining PR patterns
- **odh-linter**: Production-ready static analysis tool

### 2. Independent Distribution

- odh-linter can be:
  - Published to GitHub as standalone repo
  - Distributed via `go install`
  - Used without the research infrastructure

### 3. Cleaner Module Structure

- Independent go.mod files
- No dependencies on analysis tools
- Standard Go project layout

## Module Path Changes

### Old Paths

```go
import "github.com/opendatahub-io/rule-creator/linters/typeassert"
```

### New Paths

```go
import "github.com/opendatahub-io/odh-linter/linters/typeassert"
```

## What Stayed in rule-creator/

The research and analysis infrastructure remains:

```
rule-creator/
├── src/
│   ├── loaders/pr_loader.py           # Load PR data
│   ├── processors/comment_filter.py    # Filter comments
│   └── analyzers/                      # Pattern analysis
├── data/
│   ├── processed/                      # Extracted comments
│   └── raw/ → ../.data/               # Symlink to scraper cache
└── docs/
    ├── RULE_FINDING_DESIGN.md          # Original design
    ├── NEW_LINTABLE_PATTERNS.md        # Pattern analysis
    ├── REFACTOR_PATTERN_ANALYSIS.md    # Refactor patterns
    ├── LINTER_RUN_RESULTS.md           # Test results
    └── COMPLETE_LINTER_SUITE.md        # Full summary
```

## Building After Migration

### Option 1: From odh-linter Directory

```bash
cd odh-linter/linters/odhlint
go build -o odhlint ./cmd/odhlint
```

### Option 2: Using go install

```bash
# After publishing to GitHub
go install github.com/opendatahub-io/odh-linter/linters/odhlint/cmd/odhlint@latest
```

## For Developers

### Adding New Linters

1. Create in odh-linter:
   ```bash
   cd odh-linter/linters
   mkdir -p newlinter/cmd/newlinter
   ```

2. Use new module path:
   ```go
   module github.com/opendatahub-io/odh-linter/linters/newlinter
   ```

3. Add to odhlint:
   ```go
   import "github.com/opendatahub-io/odh-linter/linters/newlinter"
   ```

### Research Workflow

The complete workflow now spans both directories:

```
1. Scrape PRs:     go-github-scraper/
2. Analyze data:   rule-creator/src/
3. Extract rules:  rule-creator/docs/
4. Build linters:  odh-linter/linters/
5. Distribute:     odh-linter/ (standalone)
```

## Future Plans

### Short Term

- [ ] Publish odh-linter to GitHub as separate repository
- [ ] Add to golangci-lint plugin registry
- [ ] Create pre-built binaries for releases

### Long Term

- [ ] Add more linters based on continued PR analysis
- [ ] Create web UI for rule documentation
- [ ] Integration with IDE plugins

## Migration Checklist

If you're updating references:

- [ ] Update import paths in code
- [ ] Update go.mod files
- [ ] Update CI/CD scripts
- [ ] Update documentation links
- [ ] Test builds in new location

## Links

### odh-linter (Production Tool)

- Main README: [`odh-linter/README.md`](README.md)
- Quick Start: [`odh-linter/QUICK_START.md`](QUICK_START.md)
- Provenance: [`odh-linter/PROVENANCE.md`](PROVENANCE.md)
- Rule Mapping: [`odh-linter/linters/RULE_ID_MAPPING.md`](linters/RULE_ID_MAPPING.md)

### rule-creator (Research Project)

- Design Document: `../rule-creator/RULE_FINDING_DESIGN.md`
- Pattern Analysis: `../rule-creator/NEW_LINTABLE_PATTERNS.md`
- Refactor Analysis: `../rule-creator/REFACTOR_PATTERN_ANALYSIS.md`
- Complete Summary: `../rule-creator/COMPLETE_LINTER_SUITE.md`

## Questions?

Open an issue in the odh-linter repository or check the documentation.

---

**Migration Date**: November 17, 2025  
**Module Path**: `github.com/opendatahub-io/odh-linter`  
**Go Version**: 1.21+

