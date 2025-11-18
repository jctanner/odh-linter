# Pattern Discovery Scripts

Python utilities to extract actionable patterns from GitHub PR review comments for linter rule creation.

---

## Overview

These scripts help you analyze PR review history to identify patterns worth automating as linter rules.

**Workflow**:
1. Use [go-github-scraper](https://github.com/jctanner/go-github-scraper) to fetch PR data
2. Use these scripts to extract and filter comments
3. Manually analyze interesting PRs with AI assistance (Cursor/Claude)
4. Build linters from identified patterns

---

## Prerequisites

### 1. Install go-github-scraper

```bash
go install github.com/jctanner/go-github-scraper/cmd/go-github-scraper@latest
```

### 2. Install Python Dependencies

```bash
cd scripts
pip install -r requirements.txt
```

### 3. Get GitHub Token

```bash
export GITHUB_TOKEN="your_personal_access_token"
```

---

## Usage

### Step 1: Scrape PR Data

```bash
# From repository root
go-github-scraper scrape pulls \
  --owner opendatahub-io \
  --repo opendatahub-operator \
  --cache-dir .data \
  --state all
```

This creates `.data/api.github.com/repos/owner/repo/pulls/NUMBER.json` files.

### Step 2: Load and Filter Comments

```bash
cd scripts

# Run the comment filter
python filter_comments.py
```

This will:
- Load all PRs from `.data/`
- Extract inline review comments
- Filter out noise (LGTM, thanks, etc.)
- Identify actionable feedback
- Show statistics

**Output**:
```
Total comments: 7549
Actionable: 3785 (50.1%)

Top categories:
  kubernetes: 1384
  error_handling: 669
  testing: 413
  
Top reviewers:
  zdtsw: 889
  lburgazzoli: 518
```

### Step 3: Manual Analysis

Now the **human-AI collaboration** begins:

1. **Browse interesting PRs**:
   ```bash
   ls .data/api.github.com/repos/opendatahub-io/opendatahub-operator/pulls/
   ```

2. **Open in Cursor** and ask AI to analyze:
   ```
   "What patterns do you see in .data/.../pulls/1898.json?"
   ```

3. **Request linter generation**:
   ```
   "Can you build a linter for that error demotion pattern?"
   ```

4. **Test and refine**:
   ```bash
   cd odh-linter
   make build
   ./cmd/odhlint -errordemote ./test-code/
   ```

---

## Scripts

### `load_prs.py`

Loads PR data from go-github-scraper's cache.

**Classes**:
- `PRLoader` - Loads JSON files
- `PullRequest` - PR data model
- `ReviewComment` - Comment data model

**Example**:
```python
from load_prs import PRLoader

loader = PRLoader(".data")
prs = loader.load_repository("opendatahub-io", "opendatahub-operator")
comments = loader.extract_all_comments(prs)

print(f"Found {len(comments)} review comments")
```

### `filter_comments.py`

Filters and categorizes review comments.

**Features**:
- Removes noise (LGTM, social comments, merge conflicts)
- Identifies actionable feedback (suggestions, bug reports, etc.)
- Categorizes by type (security, performance, testing, etc.)
- Detects bot vs human comments
- Extracts keywords

**Example**:
```python
from filter_comments import CommentFilter

filter = CommentFilter(exclude_bots=False)
filtered = filter.filter_comments(comments)
actionable = filter.get_actionable_comments(filtered)

for fc in actionable:
    print(f"{fc.comment.reviewer.login}: {fc.comment.body[:100]}")
```

---

## What Makes a Good Linter Pattern?

### ✅ Look For

1. **Recurring Feedback** - Same issue mentioned multiple times
2. **Debates** - Disagreements between reviewers (e.g., fail-fast vs resilient)
3. **Refactoring PRs** - Code structure changes reveal architectural patterns
4. **Specific Reviewers** - Some reviewers catch specific types of issues

### ❌ Avoid

1. **Project-Specific** - Only applies to one codebase
2. **One-Off Issues** - Happened once, unlikely to repeat
3. **Subjective Style** - Personal preference without technical merit
4. **Too Broad** - Can't be automated with AST analysis

---

## Real Examples

### Example 1: Error Demotion (ODH-ARCH-006)

**Source**: PR #1898

**Pattern**: CodeRabbit suggested logging errors instead of returning them, zdtsw questioned if that's safe.

**Discovery**:
```bash
# Human spotted debate in PR #1898
# AI analyzed: "There's disagreement about error handling..."
# AI generated linter: errordemote
```

**Result**: Linter that flags when errors are caught but only logged

### Example 2: Architectural Patterns (5 linters)

**Source**: PR #2769 (refactoring by zdtsw)

**Pattern**: PR had NO review comments but moved code between packages

**Discovery**:
```bash
# Human: "Analyze this refactor PR"
# AI: "Code moved from internal/ to pkg/, tests relocated..."
# AI generated 5 linters from file movements
```

**Result**: clusterconfig, depdirection, testlocation, boolforenum, dupedef

---

## Tips for Finding Patterns

### 1. Look at Prolific Reviewers

```python
# Find reviewers with most comments
from collections import Counter
reviewers = [c.reviewer.login for c in comments]
top_reviewers = Counter(reviewers).most_common(10)
```

Focus on PRs reviewed by these people—they likely catch patterns.

### 2. Search for Keywords

```bash
# Find all comments mentioning "error"
grep -r "error" .data/api.github.com/repos/*/pulls/*.json
```

### 3. Look at Refactoring PRs

```bash
# Find PRs with "refactor" in title
jq '.github_data.title' .data/.../pulls/*.json | grep -i refactor
```

These often reveal architectural decisions worth codifying.

### 4. Analyze Bot Comments

Don't exclude bot comments! CodeRabbit and other AI reviewers can identify valuable patterns.

---

## Statistics from opendatahub-operator

```
PRs Analyzed: 2,509
Total Comments: 7,549
Actionable: 3,785 (50%)

Linters Built: 13
Success Rate: 0.5% (13 linters from 2,509 PRs)
```

**Key Insight**: You don't need to analyze ALL comments. Focus on interesting PRs and debates.

---

## Integration with odh-linter

Once you've identified a pattern:

1. **Build the linter** (with AI assistance)
2. **Add to linters/** directory
3. **Register in cmd/odhlint/main.go**
4. **Update RULE_ID_MAPPING.md**
5. **Document provenance** (which PR inspired it)

See [RULE_CONTRIBUTION.md](../RULE_CONTRIBUTION.md) for full workflow.

---

## Troubleshooting

### "No PRs found"

- Check `.data/` directory structure
- Verify go-github-scraper ran successfully
- Ensure cache dir path is correct

### "Import errors"

```bash
# Ensure you're in scripts/ directory
cd scripts
python filter_comments.py
```

### "Rate limit errors"

The scraper handles this automatically, but if you're re-scraping:
```bash
# Wait an hour or use a different token
go-github-scraper scrape pulls --full  # Full resync
```

---

## Further Reading

- **[RULE_CONTRIBUTION.md](../RULE_CONTRIBUTION.md)** - Complete methodology
- **[PROVENANCE.md](../PROVENANCE.md)** - Where existing rules came from
- **[go-github-scraper](https://github.com/jctanner/go-github-scraper)** - Data collection tool

---

## Contributing

Found a great pattern? Submit a PR!

1. Add the new linter to `linters/`
2. Document which PR inspired it
3. Include test cases
4. Update `RULE_ID_MAPPING.md`

Happy pattern hunting! 🔍

