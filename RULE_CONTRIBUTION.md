# Rule Contribution Guide

How to extract linter rules from GitHub PR review comments using a hybrid human-AI workflow.

---

## Overview

The ODH Linter rules were created through a **three-stage hybrid process**:

1. **Data Collection**: GitHub scraper fetches all PR data
2. **Comment Extraction**: Python scripts parse and filter comments
3. **Rule Creation**: Human + AI (Cursor/Claude) identify patterns and build linters

This document explains how to replicate this process for finding new rules.

---

## Stage 1: Data Collection (Automated)

### Tool: `go-github-scraper`

**Repository**: [https://github.com/jctanner/go-github-scraper](https://github.com/jctanner/go-github-scraper)

A lightweight Go-based GitHub scraper that fetches all pull requests, issues, and their complete history with aggressive disk caching.

```bash
# Install
go install github.com/jctanner/go-github-scraper/cmd/go-github-scraper@latest

# Or build from source
git clone https://github.com/jctanner/go-github-scraper.git
cd go-github-scraper
go build -o bin/go-github-scraper ./cmd/go-github-scraper

# Configure
export GITHUB_TOKEN="your_token_here"

# Scrape all PRs from a repository
go-github-scraper scrape pulls \
  --owner opendatahub-io \
  --repo opendatahub-operator \
  --cache-dir .data

# This fetches:
# - All PR metadata
# - All review comments (inline and general)
# - All file changes with diffs
# - PR merge/close status
# - CI/workflow data (optional)
```

**Output**: JSON files stored in `.data/api.github.com/repos/owner/repo/pulls/NUMBER.json`

### What Gets Captured

Each PR JSON contains:
- **Basic metadata**: Title, description, author, dates
- **Review comments**: Inline code comments with file/line context
- **File changes**: Diffs showing what code changed
- **Reviews**: Approval/comment/request-changes events
- **CI/workflow data**: GitHub Actions results (optional)

### Key Features of go-github-scraper

Per the [repository](https://github.com/jctanner/go-github-scraper):

- ✅ **Complete PR and Issue Scraping**: All data including reviews, comments, commits
- ✅ **Efficient Disk Caching**: JSON files organized by `github.com/owner/repo`
- ✅ **Smart Incremental Updates**: Only fetch changed items (saves API calls)
- ✅ **Rate Limiting**: Auto-wait for GitHub's 5,000/hour limit
- ✅ **GitHub Enterprise Support**: Configure custom API endpoints

### Why This Matters

Having **full PR history** allows you to:
- See patterns in code review discussions
- Correlate comments with actual code changes
- Identify recurring feedback themes
- Track which reviewers catch which issues
- Analyze debates and disagreements (where rules are born!)

---

## Stage 2: Comment Extraction (Python)

### Tool: `rule-creator/src/loaders/pr_loader.py`

```python
from src.loaders.pr_loader import PRLoader

# Load all PRs
loader = PRLoader(".data")
prs = loader.load_repository("opendatahub-io", "opendatahub-operator")

# Extract all inline review comments
comments = loader.extract_all_comments(prs)

print(f"Found {len(comments)} review comments")
```

### Tool: `rule-creator/src/processors/comment_filter.py`

```python
from src.processors.comment_filter import CommentFilter

# Create filter (optionally exclude bots)
filter = CommentFilter(exclude_bots=False)

# Filter comments
filtered = filter.filter_comments(comments)

# Get only actionable feedback
actionable = filter.get_actionable_comments(filtered)

print(f"Actionable comments: {len(actionable)}")
```

### What Gets Filtered

The Python script removes:
- Empty comments
- Social comments ("LGTM", "thanks", "+1")
- Merge conflict messages
- Test failure notifications

And identifies:
- **Actionable feedback**: Contains "should", "must", "avoid", "consider"
- **Code suggestions**: Has markdown code blocks
- **Pattern discussions**: Mentions specific code structures
- **Questions that imply changes**: "why not", "have you considered"

### Why This Matters

Raw PR data has **lots of noise**. Filtering focuses on:
- Comments that suggest code changes
- Comments that identify anti-patterns
- Comments that enforce best practices

---

## Stage 3: Rule Creation (Human + AI)

This is where **Cursor + Claude** come in.

### Workflow

#### 3.1 Load Comments into AI Context

```bash
# In Cursor, open the project
cd /path/to/workspace

# AI (Claude via Cursor) can now read:
# - .data/api.github.com/repos/.../pulls/*.json
# - Extracted comment data from Python scripts
```

#### 3.2 Manual PR Review (Human-Guided)

The human developer:
1. **Identifies interesting PRs** (by scanning titles, reviewers, file changes)
2. **Asks AI to analyze specific PRs**
3. **Requests pattern extraction**

**Example conversation**:
```
Human: "Look at .data/.../pulls/1898.json - what patterns do you see?"

AI: "I see a debate between CodeRabbit and zdtsw about error handling.
     CodeRabbit suggests logging errors, zdtsw questions if that's safe."
     
Human: "Can we make a linter for that pattern?"

AI: "Yes! This is the 'error demotion' pattern..."
     [AI builds the linter]
```

#### 3.3 AI-Assisted Pattern Identification

The AI (Claude via Cursor):
1. **Reads PR JSON files** directly from `.data/`
2. **Identifies code patterns** from review comments
3. **Suggests lintable rules**
4. **Generates AST-based analyzers** in Go

#### 3.4 Human Review & Approval

The human developer:
1. **Reviews AI suggestions**
2. **Approves or rejects patterns**
3. **Requests refinements**
4. **Tests the linters**

---

## Real Examples from ODH Linter

### Example 1: Error Demotion (ODH-ARCH-006)

**Source**: PR #1898 (FIPS enablement)

**Discovery process**:
```
Human: "Is there anything interesting in pulls/1898.json?"

AI: [Analyzes PR]
    "Yes! There's a debate about error handling between CodeRabbit 
     and zdtsw. CodeRabbit suggested resilient error handling, 
     zdtsw questioned if errors should be hidden."
     
Human: "Can you build a linter for that?"

AI: [Creates errordemote linter]
    "Done! ODH-ARCH-006 flags when errors are caught but only logged."
```

**Result**: `errordemote` linter that enforces fail-fast by default, requires justification for resilience.

---

### Example 2: Architectural Patterns (ODH-ARCH-001 through 005)

**Source**: PR #2769 (refactoring by zdtsw)

**Discovery process**:
```
Human: "Look at pulls/2769.json - it's a refactor PR without comments. 
        Can we extract patterns from what zdtsw decided to change?"

AI: [Analyzes file movements and code structure changes]
    "This PR moves code from service layers to common packages.
     I can identify 5 architectural patterns:
     1. Cluster config access should be centralized
     2. pkg/ shouldn't import internal/
     3. Duplicate type definitions across packages
     4. Tests in wrong packages
     5. Boolean returns for multi-value checks"
     
Human: "Build linters for all of those except duplicates"

AI: [Creates 5 linters]
    "Done! clusterconfig, depdirection, testlocation, boolforenum"
```

**Result**: 5 architectural linters extracted from refactoring decisions.

---

## The Hybrid Approach

### What Automation Does

- ✅ **Data Collection**: GitHub API scraping (100% automated)
- ✅ **Comment Extraction**: Python parsing & filtering (100% automated)
- ✅ **Linter Code Generation**: AST analysis in Go (AI-assisted)

### What Humans Do

- ✅ **PR Selection**: Identify interesting patterns (human judgment)
- ✅ **Pattern Validation**: Confirm patterns are worth linting (human judgment)
- ✅ **Context Understanding**: Interpret debates & trade-offs (human judgment)
- ✅ **Rule Refinement**: Test and tune linters (human testing)

### What AI Does

- ✅ **JSON Parsing**: Read & understand PR structure (AI analysis)
- ✅ **Pattern Recognition**: Identify lintable patterns (AI analysis)
- ✅ **Code Generation**: Write Go AST analyzers (AI generation)
- ✅ **Documentation**: Generate READMEs and examples (AI writing)

---

## Key Insights

### 1. Not All Comments Become Rules

From **3,785 review comments**, we extracted:
- **~30 actionable patterns**
- **13 implemented linters**

Most comments are:
- Project-specific
- One-off issues
- Not generalizable

### 2. Bot Comments Can Be Valuable

CodeRabbit (AI code reviewer) contributed to the `errordemote` linter by:
- Suggesting resilient error handling
- Sparking debate with human reviewer
- Documenting trade-offs

Don't automatically exclude bot comments!

### 3. Refactoring PRs Are Gold Mines

PR #2769 (refactoring) had **zero review comments** but yielded **5 linters** by analyzing:
- What code was moved
- How structure changed
- What abstractions were created

Refactoring captures **implicit architectural decisions**.

### 4. The Best Rules Come From Debates

The `errordemote` linter exists because **two reviewers disagreed**:
- CodeRabbit: "Be resilient"
- zdtsw: "Should we hide errors?"

Disagreement = Decision point = Lintable pattern!

---

## How to Contribute New Rules

### Step 1: Find Patterns

```bash
# Install and run the scraper (if not already done)
# See: https://github.com/jctanner/go-github-scraper
go install github.com/jctanner/go-github-scraper/cmd/go-github-scraper@latest

export GITHUB_TOKEN="your_token"
go-github-scraper scrape pulls --owner ORG --repo REPO --cache-dir .data

# Extract comments
cd rule-creator
python3 src/loaders/pr_loader.py
```

### Step 2: Manual Review

```bash
# Open in Cursor
cursor /path/to/workspace

# Browse interesting PRs
ls .data/api.github.com/repos/ORG/REPO/pulls/

# Ask AI to analyze
"What patterns do you see in pulls/XXXX.json?"
```

### Step 3: Request Linter

```
"Can you build a linter for [pattern description]?"
```

### Step 4: Test & Refine

```bash
# Build the linter
cd odh-linter
make build

# Test it
./cmd/odhlint -LINTERNAME ./test-code/

# Iterate with AI
"The linter has false positives in [scenario]. Can you fix?"
```

### Step 5: Submit PR

```bash
# Add linter to unified tool
# Update documentation
# Create PR to odh-linter
```

---

## Tools & Infrastructure

### Required Tools

1. **[go-github-scraper](https://github.com/jctanner/go-github-scraper)**: Fetch PR data from GitHub API
2. **Python 3.8+**: Comment extraction and filtering
3. **Go 1.21+**: Linter implementation
4. **Cursor + Claude**: AI-assisted analysis and code generation

### Recommended Workflow

```bash
# Terminal 1: Data collection
# See: https://github.com/jctanner/go-github-scraper
export GITHUB_TOKEN="your_token"
go-github-scraper scrape pulls --owner ORG --repo REPO --cache-dir .data

# Terminal 2: Python analysis
cd rule-creator
python3 src/processors/comment_filter.py

# Terminal 3: Cursor/Claude
cursor .
# Use AI to analyze and build
```

---

## Statistics

### ODH Linter (Current)

| Metric | Value |
|--------|-------|
| **PRs Analyzed** | 3,785 |
| **Review Comments** | ~12,000 |
| **Actionable Comments** | ~3,000 |
| **Linters Built** | 13 |
| **Lines of Linter Code** | ~3,500 |
| **Human Hours** | ~40 |
| **AI Assistance** | 70% of code generation |

### Hit Rate

- **Comments → Patterns**: 3,000 → 30 (1%)
- **Patterns → Linters**: 30 → 13 (43%)
- **Success Rate**: 13 linters from 3,785 PRs (0.3%)

But those 13 linters catch **real bugs** in production code! 🎯

---

## Best Practices

### ✅ Do

- **Include bot comments** in analysis
- **Look for debates** between reviewers
- **Analyze refactoring PRs** (even without comments)
- **Test linters on real code** before finalizing
- **Document provenance** (which PR inspired the rule)

### ❌ Don't

- **Auto-generate rules** from all comments (too noisy)
- **Skip manual review** (AI needs human judgment)
- **Ignore context** (why was this comment made?)
- **Create overly strict rules** (allow suppression with //nolint)
- **Forget to update documentation** (explain the rule's origin)

---

## Future Enhancements

### Planned Features

1. **Automated pattern clustering**: Group similar comments
2. **Frequency analysis**: Find most common feedback
3. **Reviewer-specific patterns**: Learn from specific reviewers (e.g., zdtsw)
4. **Cross-repository learning**: Apply patterns from other projects

### Research Questions

- Can LLMs suggest new patterns without human guidance?
- How to automatically validate linter accuracy?
- Can we predict which comments will become rules?

---

## Conclusion

The ODH Linter methodology is a **hybrid human-AI workflow**:

1. 🤖 **Automation** handles data collection and filtering
2. 👤 **Humans** identify interesting patterns and provide judgment
3. 🤖 **AI** generates linter code and documentation
4. 👤 **Humans** test, refine, and approve

This approach combines:
- **Scale** of automation
- **Judgment** of humans  
- **Speed** of AI code generation
- **Quality** of manual review

The result: **High-value linters** that capture **real architectural decisions** from **actual code review debates**. 🎯

---

## Related Repositories

### Core Tools

- **[go-github-scraper](https://github.com/jctanner/go-github-scraper)** - GitHub PR/issue scraper with disk caching
- **[odh-linter](https://github.com/opendatahub-io/odh-linter)** - The linter suite (this project)

### Data Source

- **[opendatahub-operator](https://github.com/opendatahub-io/opendatahub-operator)** - Source of PR review data

---

## Contact & Contributions

To contribute new rules:
1. Fork the `odh-linter` repository
2. Follow this guide to extract patterns
3. Submit a PR with your new linter
4. Include provenance (which PR inspired it)

For questions:
- Open an issue in the repository
- Reference this guide
- Include PR numbers for context

Happy linting! 🚀